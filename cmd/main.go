//go:build windows

package main

import (
	"context"
	"kakaotalkadblock/internal"
	"kakaotalkadblock/internal/win"
	"kakaotalkadblock/winres"
	"os/exec"

	"github.com/energye/systray"

	_ "kakaotalkadblock/winres"
)

const VERSION = "2.2.0"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	destroy := func() {
		cancel()
		systray.Quit()
	}
	systray.SetOnRClick(func(menu systray.IMenu) {
		menu.ShowMenu()
	})
	systray.Run(func() {
		systray.SetIcon(winres.IconData)
		systray.SetTooltip("KakaoTalkAdBlock")
		systray.AddMenuItem(VERSION, VERSION).Disable()
		systray.AddMenuItem("Check latest releases", "Check latest releases").Click(func() {
			exec.Command(
				"rundll32",
				"url.dll,FileProtocolHandler",
				"https://github.com/blurfx/KakaoTalkAdBlock/releases",
			).Start()
		})
		systray.AddSeparator()

		startupItem := systray.AddMenuItem("Run on startup", "Run on startup")
		if win.IsStartupEnabled() {
			startupItem.Check()
		}
		startupItem.Click(func() {
			if startupItem.Checked() {
				startupItem.Uncheck()
				win.SetStartupEnabled(false)
			} else {
				startupItem.Check()
				win.SetStartupEnabled(true)
			}
		})
		systray.AddMenuItem("E&xit", "Exit").Click(destroy)

		internal.Run(ctx)
	}, func() {
		cancel()
	})
}
