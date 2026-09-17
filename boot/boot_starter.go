package boot

import (
	"fmt"
	"os"
)

// configureStarters 调用所有匹配的启动器的 Configure 方法。
func (b *Boot) configureStarters() error {
	for _, s := range b.starters {
		if !b.starterMatches(s) {
			continue
		}
		if err := s.Configure(newAppCtx(b.ctx, b.rootCtx)); err != nil {
			return b.reportError("初始化", fmt.Errorf("启动器 %s 配置失败: %w", s.Name(), err))
		}
	}
	return nil
}

// startStarters 启动所有匹配的启动器，部分失败时逆序停止已启动的启动器。
func (b *Boot) startStarters(report *StartupReport) error {
	started := make([]Starter, 0, len(b.starters))
	starterCount := 0
	for _, s := range b.starters {
		if !b.starterMatches(s) {
			continue
		}
		if err := s.Start(newAppCtx(b.ctx, b.rootCtx)); err != nil {
			// 逆序停止已启动的 Starter，避免部分启动失败导致资源泄漏
			for i := len(started) - 1; i >= 0; i-- {
				if stopErr := started[i].Stop(newAppCtx(b.ctx, b.rootCtx)); stopErr != nil {
					fmt.Fprintf(os.Stderr, "starter %s stop error: %v\n", started[i].Name(), stopErr)
				}
			}
			return b.reportError("初始化", fmt.Errorf("启动器 %s 启动失败: %w", s.Name(), err))
		}
		started = append(started, s)
		starterCount++
		report.AddStarter(s.Name())
	}
	report.SetStarterCount(starterCount)
	b.startersStarted = true
	return nil
}
