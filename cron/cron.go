package cron

import (
	"context"
	"fmt"
	"host-agent/logger"

	"time"

	"github.com/go-co-op/gocron/v2"
)

//var Cron gocron.Scheduler

// 定时任务管理器
type sTimer struct {
	cron gocron.Scheduler
}

var corn *sTimer

func init() {
	logger.Info("创建注册定时器完成")
	corn = New()
}

func New() *sTimer {

	// 加载时区
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}

	s, err := gocron.NewScheduler(
		gocron.WithLocation(location),
	)
	timer := &sTimer{
		cron: s,
	}
	if err != nil {
		logger.Error("无法创建定时器")
		panic(err)
	}

	//开始运行定时器
	timer.cron.Start()

	return timer
}

// interval单位：秒
func (s *sTimer) Add(tag string, interval time.Duration, f func(), exec bool) error {
	_, err := s.cron.NewJob(
		// gocron.DurationJob(time.Duration(hour)*time.Hour),
		gocron.DurationJob(interval),
		gocron.NewTask(f),
		gocron.WithTags(tag),
	)
	if err != nil {
		logger.Error(fmt.Sprintf("%v", err))
		return err
	}
	if exec {
		f()
	}
	return nil
}

// AddStartAt 首次执行时间，与Add 不同的是 Add是立刻执行，本函数可以指定首次执行的时间，时间必须是未来的时间
func (s *sTimer) AddStartAt(ctx context.Context, tag string, interval int, timeAt time.Time, f func()) error {
	_, err := s.cron.NewJob(
		gocron.DurationJob(time.Duration(interval)*time.Hour*24),
		gocron.NewTask(f),
		gocron.WithTags(tag),
		gocron.WithStartAt(gocron.WithStartDateTime(timeAt)),
	)
	if err != nil {
		logger.Error(fmt.Sprintf("%v", err))
		return err
	}
	return nil
}

// AddStartTimestampAt timeAt:首次执行时间毫秒时间戳，时间必须是未来的时间
func (s *sTimer) AddStartTimestampAt(ctx context.Context, tag string, interval int, timeAt int64, f func()) error {
	now := time.Now().UnixMilli()
	if timeAt < now {
		for {
			timeAt = timeAt + int64(interval*24*3600*1000)
			if timeAt > now {
				break
			}
		}
	}

	startTime := time.UnixMilli(timeAt)
	_, err := s.cron.NewJob(
		gocron.DurationJob(time.Duration(interval)*time.Hour*24),
		gocron.NewTask(f),
		gocron.WithTags(tag),
		gocron.WithStartAt(gocron.WithStartDateTime(startTime)),
	)
	if err != nil {
		logger.Error(fmt.Sprintf("%v", err))
		return err
	}
	return nil
}

func (s *sTimer) Remove(tag string) {
	s.cron.RemoveByTags(tag)
}

/**
 * @Description: 添加定时任务
 * @param name 任务名称
 * @param cronTime 定时时间
 * @param f 任务函数
 * @param exec 是否立即执行
 * @return error
 */
func (s *sTimer) CronJob(name string, cronTime string, f func(), exec bool) error {
	_, err := s.cron.NewJob(
		gocron.CronJob(cronTime, false),
		gocron.NewTask(
			f,
		),
		gocron.WithTags(name),
	)

	if err != nil {
		return err
	}

	if exec {
		f()
	}
	return nil

}

func GetCron() *sTimer {
	return corn
}
