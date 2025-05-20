{{/*
Expand the name of the chart.
*/}}
{{- define "shorty.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "api.name" -}}
{{ include "shorty.name" . }}-api
{{- end }}

{{- define "web.name" -}}
{{ include "shorty.name" . }}-web
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "shorty.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{- define "api.fullname" -}}
{{ include "shorty.fullname" . | trunc 60 | trimSuffix "-"}}-api
{{- end }}

{{- define "web.fullname" -}}
{{ include "shorty.fullname" . | trunc 60 | trimSuffix "-"}}-web
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "shorty.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "shorty.labels" -}}
helm.sh/chart: {{ include "shorty.chart" . }}
{{ include "shorty.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "api.labels" -}}
helm.sh/chart: {{ include "shorty.chart" . }}
{{ include "api.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "web.labels" -}}
helm.sh/chart: {{ include "shorty.chart" . }}
{{ include "web.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "shorty.selectorLabels" -}}
app.kubernetes.io/name: {{ include "shorty.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "api.selectorLabels" -}}
app.kubernetes.io/name: {{ include "shorty.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}-api
{{- end }}

{{- define "web.selectorLabels" -}}
app.kubernetes.io/name: {{ include "shorty.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}-web
{{- end }}


{{/*
Create the name of the service account to use
*/}}
{{- define "api.serviceAccountName" -}}
{{- if .Values.server.serviceAccount.create }}
{{- default (include "shorty.fullname" .) .Values.server.serviceAccount.name }}-api
{{- else }}
{{- default "default" .Values.server.serviceAccount.name }}-api
{{- end }}
{{- end }}

{{- define "web.serviceAccountName" -}}
{{- if .Values.web.serviceAccount.create }}
{{- default (include "shorty.fullname" .) .Values.server.serviceAccount.name }}-web
{{- else }}
{{- default "default" .Values.server.serviceAccount.name }}-web
{{- end }}
{{- end }}


{{/*
Create the url of the application
*/}}
{{- define "api.url" -}}
{{- if .Values.api.ingress.tls -}}
https://{{- .Values.api.hostname }}
{{- else }}
http://{{- .Values.api.hostname }}
{{- end }}
{{- end }}
