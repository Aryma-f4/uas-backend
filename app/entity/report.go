package entity

type StatisticsResponse struct {
	TotalAchievements    int                    `json:"total_achievements"`
	TotalVerified        int                    `json:"total_verified"`
	TotalPending         int                    `json:"total_pending"`
	TotalRejected        int                    `json:"total_rejected"`
	ByType               map[string]int         `json:"by_type"`
	ByStatus             map[string]int         `json:"by_status"`
	ByCompetitionLevel   map[string]int         `json:"by_competition_level,omitempty"`
	TopStudents          []TopStudentStats      `json:"top_students,omitempty"`
	MonthlyTrend         []MonthlyStats         `json:"monthly_trend,omitempty"`
}

type TopStudentStats struct {
	StudentID    string `json:"student_id"`
	StudentName  string `json:"student_name"`
	TotalPoints  int    `json:"total_points"`
	Achievements int    `json:"achievements"`
}

type MonthlyStats struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

type StudentReportResponse struct {
	StudentInfo     StudentReportInfo    `json:"student_info"`
	Statistics      StatisticsResponse   `json:"statistics"`
	Achievements    []AchievementResponse `json:"achievements"`
}

type StudentReportInfo struct {
	ID           string `json:"id"`
	StudentID    string `json:"student_id"`
	FullName     string `json:"full_name"`
	ProgramStudy string `json:"program_study"`
	AcademicYear string `json:"academic_year"`
	AdvisorName  string `json:"advisor_name,omitempty"`
}
