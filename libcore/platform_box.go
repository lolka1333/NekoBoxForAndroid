package libcore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"libcore/procfs"
	"log"
	"net/netip"
	"strings"
	"syscall"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/process"
	"github.com/sagernet/sing-box/experimental/libbox/platform"
	"github.com/sagernet/sing-box/option"
	tun "github.com/sagernet/sing-tun"
	"github.com/sagernet/sing/common/control"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/logger"
	N "github.com/sagernet/sing/common/network"
)

var boxPlatformInterfaceInstance platform.Interface = &boxPlatformInterfaceWrapper{}

type boxPlatformInterfaceWrapper struct{}

func (w *boxPlatformInterfaceWrapper) ReadWIFIState() adapter.WIFIState {
	state := strings.Split(intfBox.WIFIState(), ",")
	return adapter.WIFIState{
		SSID:  state[0],
		BSSID: state[1],
	}
}

func (w *boxPlatformInterfaceWrapper) Initialize(ctx context.Context, router adapter.Router) error {
	return nil
}

func (w *boxPlatformInterfaceWrapper) UsePlatformAutoDetectInterfaceControl() bool {
	return true
}

func (w *boxPlatformInterfaceWrapper) AutoDetectInterfaceControl() control.Func {
	return func(network, address string, conn syscall.RawConn) error {
		return control.Raw(conn, func(fd uintptr) error {
			return intfBox.AutoDetectInterfaceControl(int32(fd))
		})
	}
}

func (w *boxPlatformInterfaceWrapper) OpenTun(options *tun.Options, platformOptions option.TunPlatformOptions) (tun.Tun, error) {
	if len(options.IncludeUID) > 0 || len(options.ExcludeUID) > 0 {
		return nil, E.New("android: unsupported uid options")
	}
	if len(options.IncludeAndroidUser) > 0 {
		return nil, E.New("android: unsupported android_user option")
	}

	optionsJSON, _ := json.Marshal(options)
	platformOptionsJSON, _ := json.Marshal(platformOptions)

	tunFd, err := intfBox.OpenTun(string(optionsJSON), string(platformOptionsJSON))
	if err != nil {
		return nil, fmt.Errorf("intfBox.OpenTun: %v", err)
	}

	tunFd, err = syscall.Dup(tunFd)
	if err != nil {
		return nil, fmt.Errorf("syscall.Dup: %v", err)
	}

	options.FileDescriptor = int(tunFd)
	return tun.New(*options)
}

func (w *boxPlatformInterfaceWrapper) CloseTun() error {
	return nil
}

func (w *boxPlatformInterfaceWrapper) UsePlatformDefaultInterfaceMonitor() bool {
	return true
}

func (w *boxPlatformInterfaceWrapper) CreateDefaultInterfaceMonitor(l logger.Logger) tun.DefaultInterfaceMonitor {
	return &interfaceMonitor{}
}

func (w *boxPlatformInterfaceWrapper) UsePlatformInterfaceGetter() bool {
	return false
}

func (w *boxPlatformInterfaceWrapper) Interfaces() ([]control.Interface, error) {
	return nil, errors.New("operation not supported")
}

func (w *boxPlatformInterfaceWrapper) IncludeAllNetworks() bool {
	return false
}

func (w *boxPlatformInterfaceWrapper) UnderNetworkExtension() bool {
	return false
}

func (w *boxPlatformInterfaceWrapper) ClearDNSCache() {}

func (w *boxPlatformInterfaceWrapper) FindProcessInfo(ctx context.Context, network string, source netip.AddrPort, destination netip.AddrPort) (*process.Info, error) {
	var uid int32
	if useProcfs {
		uid = procfs.ResolveSocketByProcSearch(network, source, destination)
		if uid == -1 {
			return nil, E.New("procfs: not found")
		}
	} else {
		ipProtocol, err := getIPProtocol(network)
		if err != nil {
			return nil, err
		}

		uid, err = intfBox.FindConnectionOwner(ipProtocol, source.Addr().String(), int32(source.Port()), destination.Addr().String(), int32(destination.Port()))
		if err != nil {
			return nil, err
		}
	}

	packageName, _ := intfBox.PackageNameByUid(uid)
	return &process.Info{UserId: uid, PackageName: packageName}, nil
}

func getIPProtocol(network string) (int32, error) {
	switch N.NetworkName(network) {
	case N.NetworkTCP:
		return syscall.IPPROTO_TCP, nil
	case N.NetworkUDP:
		return syscall.IPPROTO_UDP, nil
	default:
		return 0, E.New("unknown network: ", network)
	}
}

var disableSingBoxLog = false

func (w *boxPlatformInterfaceWrapper) Write(p []byte) (n int, err error) {
	if !disableSingBoxLog {
		log.Print(string(p))
	}
	return len(p), nil
}
