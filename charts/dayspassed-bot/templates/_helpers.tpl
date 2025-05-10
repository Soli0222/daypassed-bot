{{/*
Expand the name of the chart.
*/}}
{{- define "dayspassed-bot.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "dayspassed-bot.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "dayspassed-bot.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "dayspassed-bot.labels" -}}
helm.sh/chart: {{ include "dayspassed-bot.chart" . }}
{{ include "dayspassed-bot.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/name: {{ include "dayspassed-bot.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "dayspassed-bot.selectorLabels" -}}
app.kubernetes.io/name: {{ include "dayspassed-bot.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Job Selector labels
*/}}
{{- define "dayspassed-bot.jobselectorLabels" -}}
app.kubernetes.io/name: {{ include "dayspassed-bot.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Name for the OnePasswordItem resource and the Kubernetes Secret it generates.
If 'itemResourceNameOverride' is set in values.yaml, it will be used.
Otherwise, a default name is constructed using the full chart name and a suffix.
*/}}
{{- define "dayspassed-bot.onepasswordItemName" -}}
{{- .Values.onepassword.itemResourceNameOverride | default (printf "%s-1password" (include "dayspassed-bot.fullname" .)) -}}
{{- end -}}
