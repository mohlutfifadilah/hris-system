package utils

import (
	"fmt"
	"html/template"
	"time"
	// atau uuid lib yang kamu pakai
)

// IndonesianDateFormat helper untuk template
func IndonesianDateFormat(t time.Time) string {
    if t.IsZero() {
        return "-"
    }

    days := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
    months := []string{"January", "February", "March", "April", "May", "June",
        "July", "August", "September", "October", "November", "December"}

    dayName := days[int(t.Weekday())]
    month := months[t.Month()-1]
    day := t.Day()
    year := t.Year()
    
    return fmt.Sprintf("%s, %d %s %d", dayName, day, month, year)
}

// RegisterTemplateHelpers register semua helper ke template
func RegisterTemplateHelpers(tmpl *template.Template) *template.Template {
    return tmpl.Funcs(template.FuncMap{
        "indDate": IndonesianDateFormat,
    })
}