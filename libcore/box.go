package libcore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/matsuridayo/libneko/neko_log"
	"github.com/matsuridayo/libneko/protect_server"
	"github.com/matsuridayo/libneko/speedtest"
	"github.com/sagernet/sing-box/boxapi"

	"libcore/device"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/common/conntrack"
	"github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/outbound"
	"github.com/sagernet/sing/service/pause"
)

var mainInstance *BoxInstance

func VersionBox() string {
	version := []string{
		"sing-box: " + constant.Version,
		fmt.Sprintf("%s@%s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
	}

	if debugInfo, loaded := debug.ReadBuildInfo(); loaded {
		for _, setting := range debugInfo.Settings {
			if setting.Key == "-tags" && setting.Value != "" {
				version = append(version, setting.Value)
			}
		}
	}

	return strings.Join(version, "\n")
}

func ResetAllConnections(system bool) {
	if system {
		conntrack.Close()
		log.Println("[Debug] Reset system connections done")
	}
}

type BoxInstance struct {
	*box.Box
	cancel       context.CancelFunc
	state        int
	v2api        *boxapi.SbV2rayServer
	selector     *outbound.Selector
	pauseManager pause.Manager
	ForTest      bool
}

func NewSingBoxInstance(config string) (*BoxInstance, error) {
	defer device.DeferPanicToError("NewSingBoxInstance", func(err error) { err = err })

	var options option.Options
	if err := options.UnmarshalJSON([]byte(config)); err != nil {
		return nil, fmt.Errorf("decode config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	sleepManager := pause.ManagerFromContext(ctx)
	ctx = pause.ContextWithManager(ctx, sleepManager)

	instance, err := box.New(box.Options{
		Options:           options,
		Context:           ctx,
		PlatformInterface: boxPlatformInterfaceInstance,
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("create service: %v", err)
	}

	b := &BoxInstance{
		Box:          instance,
		cancel:       cancel,
		pauseManager: sleepManager,
	}

	b.SetLogWritter(neko_log.LogWriter)

	pf := instance.GetLogPlatformFormatter()
	pf.DisableColors = true
	pf.DisableLineBreak = false

	if proxy, ok := b.Router().Outbound("proxy"); ok {
		if selector, ok := proxy.(*outbound.Selector); ok {
			b.selector = selector
		}
	}

	return b, nil
}

func (b *BoxInstance) Start() error {
	defer device.DeferPanicToError("box.Start", func(err error) { err = err })

	if b.state == 0 {
		b.state = 1
		return b.Box.Start()
	}
	return errors.New("already started")
}

func (b *BoxInstance) Close() error {
	defer device.DeferPanicToError("box.Close", func(err error) { err = err })

	if b.state == 2 {
		return nil
	}
	b.state = 2

	if mainInstance == b {
		mainInstance = nil
		goServeProtect(false)
	}

	b.Close()
	b.Box.Close()

	return nil
}

func (b *BoxInstance) Sleep() {
	b.pauseManager.DevicePause()
	_ = b.Box.Router().ResetNetwork()
}

func (b *BoxInstance) Wake() {
	b.pauseManager.DeviceWake()
}

func (b *BoxInstance) SetAsMain() {
	mainInstance = b
	goServeProtect(true)
}

func (b *BoxInstance) SetConnectionPoolEnabled(enable bool) {
	// TODO: Implement API
}

func (b *BoxInstance) SetV2rayStats(outbounds string) {
	b.v2api = boxapi.NewSbV2rayServer(option.V2RayStatsServiceOptions{
		Enabled:   true,
		Outbounds: strings.Split(outbounds, "\n"),
	})
	b.Box.Router().SetV2RayServer(b.v2api)
}

func (b *BoxInstance) QueryStats(tag, direct string) int64 {
	if b.v2api == nil {
		return 0
	}
	return b.v2api.QueryStats(fmt.Sprintf("outbound>>>%s>>>traffic>>>%s", tag, direct))
}

func (b *BoxInstance) SelectOutbound(tag string) bool {
	return b.selector != nil && b.selector.SelectOutbound(tag)
}

func UrlTest(i *BoxInstance, link string, timeout int32) (int32, error) {
	defer device.DeferPanicToError("box.UrlTest", func(err error) { err = err })

	client := boxapi.CreateProxyHttpClient(mainInstance.Box)
	if i != nil {
		client = boxapi.CreateProxyHttpClient(i.Box)
	}
	return speedtest.UrlTest(client, link, timeout, speedtest.UrlTestStandard_RTT)
}

var protectCloser io.Closer

func goServeProtect(start bool) {
	if protectCloser != nil {
		protectCloser.Close()
		protectCloser = nil
	}
	if start {
		protectCloser = protect_server.ServeProtect("protect_path", false, 0, func(fd int) {
			intfBox.AutoDetectInterfaceControl(int32(fd))
		})
	}
}
