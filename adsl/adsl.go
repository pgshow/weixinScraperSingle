package adsl

import (
	"strings"
	"syscall"
	"time"
	"weixinScraperSingle/gocommand"
	"weixinScraperSingle/util"
)

var (
	logger     = util.GetLogger("adsl")
	ipChanging = new(bool)
)

func ChangeIP() {
	if *ipChanging == true {
		goto adslEnd
	}

	*ipChanging = true

	if pppoeStop() {
		for i := 1; i <= 8; i++ {
			time.Sleep(1 * time.Second)
			if pppoeStatus() == "off" {
				// 等待直到断开拨号
				break
			}
		}

		time.Sleep(2 * time.Second)

		if pppoeStart() {
			for i := 1; i <= 8; i++ {
				time.Sleep(1 * time.Second)
				if pppoeStatus() == "on" {
					// 等待直到拨号成功
					time.Sleep(2 * time.Second)
					break
				}
			}
		}
	}

	*ipChanging = false

adslEnd:
}

func pppoeStart() (result bool) {
	_, success := cmd("/usr/sbin/pppoe-start")

	if success == false {
		time.Sleep(time.Second)
		return
	}

	logger.Info("pppoe start")
	return true
}

func pppoeStop() (result bool) {
	_, success := cmd("/usr/sbin/pppoe-stop")

	if success == false {
		time.Sleep(time.Second)
		return
	}

	logger.Info("pppoe stop")
	return true

}

func pppoeStatus() (status string) {
	out, success := cmd("/usr/sbin/pppoe-status")
	if success == false {
		logger.Error("pppoe-status failed")
		return
	}

	if strings.Contains(out, "Link is up and running") {
		return "on"
	} else if strings.Contains(out, "Link is down") {
		return "off"
	}

	logger.Errorf("pppoe-status return unexpect: ", out)
	if strings.Contains(out, "ppp2 is down") || strings.Contains(out, "Cannot find") {
		pppoeStop()
		time.Sleep(time.Minute)
	}
	return
}

func cmd(command string) (result string, success bool) {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("Command 发生严重错误", err)
		}
	}()

	var cmd, out, err = gocommand.NewCommand().Exec(command)

	if cmd != nil {
		pgid, err := syscall.Getpgid(cmd.Process.Pid)
		if err == nil {
			errKill := syscall.Kill(-pgid, 15) // note the minus sign
			if errKill != nil {
				logger.Errorf("Kill 发生错误", errKill)
			}
		}

		_ = cmd.Wait()
	}

	if err != nil {
		logger.Errorf("执行命令 %s, 发生错误 %s", command, err)
		return
	}

	return out, true
}

func WaitChangingIp() {
	if *ipChanging {
		// wait if the adsl is changing ip
		for {
			if !*ipChanging {
				break
			}
			time.Sleep(time.Second)
		}
	}
}
