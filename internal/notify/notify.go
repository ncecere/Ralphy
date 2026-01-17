package notify

import (
	"os/exec"
	"runtime"
)

func Done(message string) {
	switch runtime.GOOS {
	case "darwin":
		playMacSound()
		showMacNotification("Ralphy", message)
	case "linux":
		playLinuxSound()
		showLinuxNotification("Ralphy", message)
	case "windows":
		playWindowsSound()
	}
}

func Error(message string) {
	switch runtime.GOOS {
	case "darwin":
		showMacNotification("Ralphy - Error", message)
	case "linux":
		showLinuxNotificationUrgent("Ralphy - Error", message)
	}
}

func playMacSound() {
	_ = exec.Command("afplay", "/System/Library/Sounds/Glass.aiff").Start()
}

func showMacNotification(title, message string) {
	script := `display notification "` + message + `" with title "` + title + `"`
	_ = exec.Command("osascript", "-e", script).Run()
}

func playLinuxSound() {
	_ = exec.Command("paplay", "/usr/share/sounds/freedesktop/stereo/complete.oga").Start()
}

func showLinuxNotification(title, message string) {
	_ = exec.Command("notify-send", title, message).Run()
}

func showLinuxNotificationUrgent(title, message string) {
	_ = exec.Command("notify-send", "-u", "critical", title, message).Run()
}

func playWindowsSound() {
	_ = exec.Command("powershell.exe", "-Command", "[System.Media.SystemSounds]::Asterisk.Play()").Run()
}
