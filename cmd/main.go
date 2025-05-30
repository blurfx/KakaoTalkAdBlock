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

const VERSION = "2.2.2"

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
		checkRelease := systray.AddMenuItem("", "Check latest releases")
		checkRelease.Click(func() {
			exec.Command(
				"rundll32",
				"url.dll,FileProtocolHandler",
				"https://github.com/blurfx/KakaoTalkAdBlock/releases",
			).Start()
		})
		checkRelease.Hide()
		systray.AddSeparator()
		go func() {
			tagName, hasNewRelease := internal.CheckLatestVersion(VERSION)
			if hasNewRelease {
				checkRelease.SetTitle("New version available: " + tagName)
				checkRelease.Show()
			}
		}()

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
