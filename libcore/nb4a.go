package libcore

import (
	"libcore/device"
	"log"
	"os"
	"path/filepath"
	"runtime"
	_ "unsafe"

	"github.com/matsuridayo/libneko/neko_common"
	"github.com/matsuridayo/libneko/neko_log"
	boxmain "github.com/sagernet/sing-box/cmd/sing-box"
)

//go:linkname resourcePaths github.com/sagernet/sing-box/constant.resourcePaths
var resourcePaths []string

func NekoLogPrintln(message string) {
	log.Println(message)
}

func NekoLogClear() {
	neko_log.LogWriter.Truncate()
}

func ForceGc() {
	go runtime.GC()
}

func InitCore(process, cachePath, internalAssets, externalAssets string,
	maxLogSizeKb int32, logEnable bool,
	if1 NB4AInterface, if2 BoxPlatformInterface,
) {
	defer device.DeferPanicToError("InitCore", func(err error) { log.Println(err) })
	isBgProcess := true //strings.HasSuffix(process, ":bg")

	neko_common.RunMode = neko_common.RunMode_NekoBoxForAndroid
	intfNB4A = if1
	intfBox = if2
	useProcfs = intfBox.UseProcFS()

	// Установка рабочего каталога
	workingDir := filepath.Join(cachePath, "../no_backup")
	if err := os.MkdirAll(workingDir, 0755); err != nil {
		log.Println("Ошибка создания каталога:", err)
	}
	if err := os.Chdir(workingDir); err != nil {
		log.Println("Ошибка смены каталога:", err)
	}

	// Настройка путей ресурсов
	resourcePaths = append(resourcePaths, externalAssets)

	// Настройка логирования
	if maxLogSizeKb < 50 {
		maxLogSizeKb = 50
	}
	neko_log.LogWriterDisable = !logEnable
	neko_log.TruncateOnStart = isBgProcess
	neko_log.SetupLog(int(maxLogSizeKb)*1024, filepath.Join(cachePath, "neko.log"))
	boxmain.SetDisableColor(true)

	// Настройка компонентов
	go func() {
		defer device.DeferPanicToError("InitCore-go", func(err error) { log.Println(err) })
		device.GoDebug(process)

		externalAssetsPath := externalAssets
		internalAssetsPath := internalAssets

		// Загрузка сертификатов
		pem, err := os.ReadFile(filepath.Join(externalAssetsPath, "ca.pem"))
		if err == nil {
			updateRootCACerts(pem)
		} else {
			log.Println("Ошибка чтения сертификата:", err)
		}

		// Обработка фонового процесса
		if isBgProcess {
			extractAssets()
		}
	}()
}
