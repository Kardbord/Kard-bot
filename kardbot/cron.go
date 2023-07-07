package kardbot

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron"

	log "github.com/sirupsen/logrus"
)

var scheduler = func() *gocron.Scheduler { return nil }

func init() {
	s := gocron.NewScheduler(time.Local)
	if s == nil {
		log.Fatal("Could not create scheduler")
	}
	scheduler = func() *gocron.Scheduler { return s }

	// https://crontab.guru/#0_9_*_*_3
	scheduler().Cron("0 9 * * 3").Do(itIsWednesdayMyDudes)

	// https://crontab.guru/#*_*_*_*_*
	scheduler().Cron("* * * * *").Do(setStatus)

	// https://crontab.guru/#30_7_*_*_*
	scheduler().Cron("30 7 * * *").Do(sendMorningCompliments)

	// https://crontab.guru/#30_20_*_*_*
	scheduler().Cron("30 20 * * *").Do(sendEveningCompliments)

	// https://crontab.guru/#0_0_*_*_*
	scheduler().Cron("0 0 * * *").Do(sendCreepyDMs)

	// https://crontab.guru/#*_*_*_*_*
	scheduler().Cron("* * * * *").Do(updateServerClocks)

	// https://crontab.guru/#0_1_*_*_*
	scheduler().Cron("0 1 * * *").Do(func() {
		if err := purgeFinishedPolls(); err != nil {
			log.Error(err)
		}
	})

	// ^The above only initializes the scheduler, it does not start it.
}

const idleTimeoutMinutes time.Duration = time.Minute * 5

func setStatus() {
	if bot().status.Load() != string(discordgo.StatusIdle) && time.Since(bot().lastActive.Load()) > idleTimeoutMinutes {
		err := bot().Session.UpdateListeningStatus("")
		if err != nil {
			log.Error(err)
		}

		idleSince := int(time.Now().Local().UnixMilli())
		err = bot().Session.UpdateStatusComplex(discordgo.UpdateStatusData{
			IdleSince: &idleSince,
			AFK:       true,
			Status:    string(discordgo.StatusIdle),
		})
		if err != nil {
			log.Error(err)
		} else {
			bot().status.Store(string(discordgo.StatusIdle))
			log.Infof("Set bot status to %s", bot().status.Load())
		}
	}
}
