package cron

import (
	"github.com/portainer/portainer"
	"github.com/robfig/cron"
)

// JobScheduler represents a service for managing crons
type JobScheduler struct {
	cron *cron.Cron
}

// NewJobScheduler initializes a new service
func NewJobScheduler() *JobScheduler {
	return &JobScheduler{
		cron: cron.New(),
	}
}

// CreateSchedule schedules the execution of a job via a runner
func (scheduler *JobScheduler) CreateSchedule(schedule *portainer.Schedule, runner portainer.JobRunner) error {
	runner.SetScheduleID(schedule.ID)
	return scheduler.cron.AddJob(schedule.CronExpression, runner)
}

// UpdateSchedule updates a specific scheduled job by re-creating a new cron
// and adding all the existing jobs. It will then re-schedule the new job
// via the specified JobRunner parameter.
// NOTE: the cron library do not support updating schedules directly
// hence the work-around
func (scheduler *JobScheduler) UpdateSchedule(schedule *portainer.Schedule, runner portainer.JobRunner) error {
	cronEntries := scheduler.cron.Entries()
	newCron := cron.New()

	for _, entry := range cronEntries {

		if entry.Job.(portainer.JobRunner).GetScheduleID() == schedule.ID {

			var jobRunner cron.Job = runner
			if entry.Job.(portainer.JobRunner).GetJobType() == portainer.SnapshotJobType {
				jobRunner = entry.Job
			}

			err := newCron.AddJob(schedule.CronExpression, jobRunner)
			if err != nil {
				return err
			}
		}

		newCron.Schedule(entry.Schedule, entry.Job)
	}

	scheduler.cron.Stop()
	scheduler.cron = newCron
	scheduler.cron.Start()
	return nil
}

// RemoveSchedule remove a scheduled job by re-creating a new cron
// and adding all the existing jobs except for the one specified via scheduleID.
// NOTE: the cron library do not support removing schedules directly
// hence the work-around
func (scheduler *JobScheduler) RemoveSchedule(scheduleID portainer.ScheduleID) {
	cronEntries := scheduler.cron.Entries()
	newCron := cron.New()

	for _, entry := range cronEntries {

		if entry.Job.(portainer.JobRunner).GetScheduleID() == scheduleID {
			continue
		}

		newCron.Schedule(entry.Schedule, entry.Job)
	}

	scheduler.cron.Stop()
	scheduler.cron = newCron
	scheduler.cron.Start()
}

// Start starts the scheduled jobs
func (scheduler *JobScheduler) Start() {
	if len(scheduler.cron.Entries()) > 0 {
		scheduler.cron.Start()
	}
}
