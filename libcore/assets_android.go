//go:build android

package libcore

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/mobile/asset"
)

func extractAssets() {
	useOfficialAssets := intfNB4A.UseOfficialAssets()

	extract := func(assetName string) {
		if err := extractAssetByName(assetName, useOfficialAssets); err != nil {
			log.Printf("Failed to extract %s: %v", assetName, err)
		}
	}

	extract(geoipDat)
	extract(geositeDat)
	extract(yacdDstFolder)
}

// Extracts assets from the APK
func extractAssetByName(assetName string, useOfficialAssets bool) error {
	// Determine if the asset is replaceable and set the directory accordingly
	replaceable := true

	var version, apkPrefix string
	switch assetName {
	case geoipDat:
		version = geoipVersion
		apkPrefix = apkAssetPrefixSingBox
	case geositeDat:
		version = geositeVersion
		apkPrefix = apkAssetPrefixSingBox
	case yacdDstFolder:
		version = yacdVersion
		replaceable = false
	}

	dir := externalAssetsPath
	if !replaceable {
		dir = internalAssetsPath
	}
	dstName := filepath.Join(dir, assetName)

	var localVersion, assetVersion string

	// Load asset version from APK
	loadAssetVersion := func() error {
		av, err := asset.Open(apkPrefix + version)
		if err != nil {
			return fmt.Errorf("failed to open version in assets: %v", err)
		}
		defer av.Close()

		b, err := io.ReadAll(av)
		if err != nil {
			return fmt.Errorf("failed to read internal version: %v", err)
		}
		assetVersion = string(b)
		return nil
	}

	if err := loadAssetVersion(); err != nil {
		return err
	}

	doExtract := shouldExtractAsset(dstName, dir, version, assetVersion, useOfficialAssets, replaceable)

	if !doExtract {
		return nil
	}

	if err := performExtraction(apkPrefix, assetName, dstName); err != nil {
		return err
	}

	return updateVersionFile(dir, version, assetVersion)
}

func shouldExtractAsset(dstName, dir, version, assetVersion string, useOfficialAssets, replaceable bool) bool {
	if _, err := os.Stat(dstName); err != nil {
		return true
	}

	if useOfficialAssets || !replaceable {
		b, err := os.ReadFile(filepath.Join(dir, version))
		if err != nil {
			_ = os.RemoveAll(version)
			return true
		}

		localVersion := string(b)
		if localVersion == "Custom" {
			return false
		}

		av, err := strconv.ParseUint(assetVersion, 10, 64)
		if err != nil {
			return assetVersion != localVersion
		}

		lv, err := strconv.ParseUint(localVersion, 10, 64)
		return err != nil || av > lv
	}

	return false
}

func performExtraction(apkPrefix, assetName, dstName string) error {
	extractXz := func(f asset.File) error {
		tmpXzName := dstName + ".xz"
		if err := extractAsset(f, tmpXzName); err == nil {
			if err = Unxz(tmpXzName, dstName); err == nil {
				os.Remove(tmpXzName)
			}
		}
		return err
	}

	extractZip := func(f asset.File, outDir string) error {
		tmpZipName := dstName + ".zip"
		if err := extractAsset(f, tmpZipName); err == nil {
			if err = Unzip(tmpZipName, outDir); err == nil {
				os.Remove(tmpZipName)
			}
		}
		return err
	}

	if f, err := asset.Open(apkPrefix + assetName + ".xz"); err == nil {
		return extractXz(f)
	} else if f, err := asset.Open("yacd.zip"); err == nil {
		os.RemoveAll(dstName)
		if err := extractZip(f, internalAssetsPath); err != nil {
			return err
		}

		matches, err := filepath.Glob(filepath.Join(internalAssetsPath, "Yacd-*"))
		if err != nil {
			return fmt.Errorf("failed to glob Yacd: %v", err)
		}
		if len(matches) != 1 {
			return fmt.Errorf("expected 1 Yacd result, found %d", len(matches))
		}
		return os.Rename(matches[0], dstName)
	}

	return nil
}

func updateVersionFile(dir, version, assetVersion string) error {
	o, err := os.Create(filepath.Join(dir, version))
	if err != nil {
		return fmt.Errorf("failed to create version file: %v", err)
	}
	defer o.Close()

	if _, err = io.WriteString(o, assetVersion); err != nil {
		return err
	}

	return nil
}

func extractAsset(i asset.File, path string) error {
	defer i.Close()
	o, err := os.Create(path)
	if err != nil {
		return err
	}
	defer o.Close()

	if _, err = io.Copy(o, i); err == nil {
		log.Printf("Extracted to %s", path)
	}
	return err
}
