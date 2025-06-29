package templates

import (
	"html/template"
	"net/http"
)

type TemplateData struct {
	Title   string
	Message string
	Data    interface{}
}

var dashboardTemplate *template.Template

func init() {
	var err error
	dashboardTemplate, err = template.New("dashboard").Parse(DashboardTemplate)
	if err != nil {
		panic("Failed to parse dashboard template: " + err.Error())
	}
}

func RenderDashboard(w http.ResponseWriter, data *TemplateData) error {
	w.Header().Set("Content-Type", "text/html")
	return dashboardTemplate.Execute(w, data)
}
