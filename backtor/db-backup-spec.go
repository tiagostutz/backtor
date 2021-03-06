package backtor

import (
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

//BackupSpec bs
type BackupSpec struct {
	Name                    string     `json:"name"`
	Enabled                 int        `json:"enabled"`
	RunningCreateWorkflowID *string    `json:"runningCreateWorkflowID,omitempty"`
	BackupCronString        *string    `json:"backupCronString,omitempty"`
	WorkerConfig            *string    `json:"workerConfig,omitempty"`
	TimeoutSeconds          *int       `json:"timeoutSeconds,omitempty"`
	FromDate                *time.Time `json:"fromDate,omitempty"`
	ToDate                  *time.Time `json:"toDate,omitempty"`
	LastUpdate              time.Time  `json:"lastUpdate,omitempty"`
	RetentionMinutely       string     `json:"retentionMinutely,omitempty"`
	RetentionHourly         string     `json:"retentionHourly,omitempty"`
	RetentionDaily          string     `json:"retentionDaily,omitempty"`
	RetentionWeekly         string     `json:"retentionWeekly,omitempty"`
	RetentionMonthly        string     `json:"retentionMonthly,omitempty"`
	RetentionYearly         string     `json:"retentionYearly,omitempty"`
}

func createBackupSpec(bs BackupSpec) error {
	stmt, err1 := db.Prepare(`INSERT INTO backup_spec (
								name, enabled, running_create_workflow,
								from_date, to_date, last_update, 
								retention_minutely, retention_hourly, retention_daily, retention_weekly, 
								retention_monthly, retention_yearly, backup_cron_string,
								worker_config, timeout_seconds
							) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?);`)
	if err1 != nil {
		return err1
	}
	_, err2 := stmt.Exec(bs.Name, bs.Enabled, bs.RunningCreateWorkflowID,
		bs.FromDate, bs.ToDate, bs.LastUpdate,
		bs.RetentionMinutely, bs.RetentionHourly, bs.RetentionDaily, bs.RetentionWeekly,
		bs.RetentionMonthly, bs.RetentionYearly, bs.BackupCronString,
		bs.WorkerConfig, bs.TimeoutSeconds)
	if err2 != nil {
		return err2
	}
	return nil
}

func updateBackupSpec(bs BackupSpec) error {
	stmt, err1 := db.Prepare(`UPDATE backup_spec SET
								name=?, enabled=?, running_create_workflow=?,
								from_date=?, to_date=?, last_update=?, 
								retention_minutely=?, retention_hourly=?, retention_daily=?, retention_weekly=?, 
								retention_monthly=?, retention_yearly=?, backup_cron_string=?,
								worker_config=?, timeout_seconds=?
							  WHERE name='` + bs.Name + `';`)
	if err1 != nil {
		return err1
	}
	resp, err2 := stmt.Exec(bs.Name, bs.Enabled, bs.RunningCreateWorkflowID,
		bs.FromDate, bs.ToDate, bs.LastUpdate,
		bs.RetentionMinutely, bs.RetentionHourly, bs.RetentionDaily, bs.RetentionWeekly,
		bs.RetentionMonthly, bs.RetentionYearly, bs.BackupCronString,
		bs.WorkerConfig, bs.TimeoutSeconds)
	if err2 != nil {
		return err2
	}

	count, err3 := resp.RowsAffected()
	if err3 != nil {
		return err3
	}
	if count == 0 {
		return fmt.Errorf("Backup spec %s doesn't exist", bs.Name)
	}

	return nil
}

func getBackupSpec(backupName string) (BackupSpec, error) {
	rows, err1 := db.Query(`SELECT 
			name, enabled, running_create_workflow,
			from_date, to_date, last_update, 
			retention_minutely, retention_hourly, retention_daily, retention_weekly, 
			retention_monthly, retention_yearly, backup_cron_string,
			worker_config, timeout_seconds
			FROM backup_spec WHERE name='` + backupName + `';`)
	if err1 != nil {
		return BackupSpec{}, err1
	}
	defer rows.Close()

	for rows.Next() {
		b := BackupSpec{}
		err2 := rows.Scan(&b.Name, &b.Enabled, &b.RunningCreateWorkflowID,
			&b.FromDate, &b.ToDate, &b.LastUpdate,
			&b.RetentionMinutely, &b.RetentionHourly, &b.RetentionDaily, &b.RetentionWeekly,
			&b.RetentionMonthly, &b.RetentionYearly, &b.BackupCronString,
			&b.WorkerConfig, &b.TimeoutSeconds)
		if err2 != nil {
			return BackupSpec{}, err2
		}
		return b, nil
	}
	err := rows.Err()
	if err != nil {
		return BackupSpec{}, err
	}
	return BackupSpec{}, fmt.Errorf("Backup spec name %s not found", backupName)
}

func listBackupSpecs(enabled *int) ([]BackupSpec, error) {
	where := ""
	if enabled != nil {
		where = fmt.Sprintf("WHERE enabled=%d", *enabled)
	}
	q := `SELECT 
			name, enabled, running_create_workflow,
			from_date, to_date, last_update, 
			retention_minutely, retention_hourly, retention_daily, retention_weekly, 
			retention_monthly, retention_yearly, backup_cron_string,
			worker_config, timeout_seconds
		FROM backup_spec ` + where + ` ORDER BY name;`

	logrus.Debugf("query=%s", q)
	rows, err1 := db.Query(q)
	if err1 != nil {
		return []BackupSpec{}, err1
	}
	defer rows.Close()

	var backups = make([]BackupSpec, 0)
	for rows.Next() {
		b := BackupSpec{}
		err2 := rows.Scan(&b.Name, &b.Enabled, &b.RunningCreateWorkflowID,
			&b.FromDate, &b.ToDate, &b.LastUpdate,
			&b.RetentionMinutely, &b.RetentionHourly, &b.RetentionDaily, &b.RetentionWeekly,
			&b.RetentionMonthly, &b.RetentionYearly, &b.BackupCronString,
			&b.WorkerConfig, &b.TimeoutSeconds)
		if err2 != nil {
			return []BackupSpec{}, err2
		}
		backups = append(backups, b)
	}
	err := rows.Err()
	if err != nil {
		return []BackupSpec{}, err
	}
	return backups, nil
}

func deleteBackupSpec(backupName string) error {
	logrus.Debugf("Deleting backup %s", backupName)
	stmt, err1 := db.Prepare(`DELETE backup_spec 
							  WHERE name='` + backupName + `';`)
	if err1 != nil {
		return err1
	}
	res, err2 := stmt.Exec()
	if err2 != nil {
		return err2
	}
	count, err3 := res.RowsAffected()
	if err3 != nil {
		return err3
	}
	logrus.Debugf("%d backup spec rows deleted", count)
	if count != 1 {
		return fmt.Errorf("Backup spec %s was not removed. count=%d", backupName, count)
	}
	return nil
}

func updateBackupSpecRunningCreateWorkflowID(backupName string, runningCreateWorkflowID *string) error {
	if runningCreateWorkflowID == nil {
		logrus.Debugf("Setting running_create_workflow of backup spec %s to nil", backupName)
	} else {
		logrus.Debugf("Setting running_create_workflow of backup spec %s to %s", backupName, *runningCreateWorkflowID)
	}

	workflowid := "NULL"
	if runningCreateWorkflowID != nil {
		workflowid = fmt.Sprintf("'%s'", *runningCreateWorkflowID)
	}
	stmt, err1 := db.Prepare(fmt.Sprintf("UPDATE backup_spec SET running_create_workflow=%s WHERE name='%s';", workflowid, backupName))
	if err1 != nil {
		return err1
	}
	res, err2 := stmt.Exec()
	if err2 != nil {
		return err2
	}

	count, err3 := res.RowsAffected()
	if err3 != nil {
		return err3
	}
	logrus.Debugf("%d backup spec rows updated", count)
	if count != 1 {
		return fmt.Errorf("running_workflowid for backup spec %s was not updated. count=%d", backupName, count)
	}
	return nil
}

func retentionParams(config string, lastReference string) []string {
	if config == "" {
		return []string{"0", lastReference}
	}
	params := strings.Split(config, "@")
	if len(params) == 1 {
		params = append(params, "L")
	}
	if params[1] == "" {
		params[1] = "L"
	}
	if params[1] == "L" {
		params[1] = lastReference
	}
	return params
}

func (b *BackupSpec) MinutelyParams() []string {
	return retentionParams(b.RetentionMinutely, "59")
}
func (b *BackupSpec) HourlyParams() []string {
	return retentionParams(b.RetentionHourly, "59")
}
func (b *BackupSpec) DailyParams() []string {
	return retentionParams(b.RetentionDaily, "23")
}
func (b *BackupSpec) WeeklyParams() []string {
	return retentionParams(b.RetentionWeekly, "7")
}
func (b *BackupSpec) MonthlyParams() []string {
	return retentionParams(b.RetentionMonthly, "L")
}
func (b *BackupSpec) YearlyParams() []string {
	return retentionParams(b.RetentionYearly, "12")
}
