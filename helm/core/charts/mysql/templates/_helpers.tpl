{{- define "higress-core-mysql.name" -}}
higress-core-mysql
{{- end -}}

{{- define "higress-core-mysql.fullname" -}}
{{- if .Values.mysql.name -}}
{{- .Values.mysql.name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-aigateway-core-mysql" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "higress-core-mysql.authSecretName" -}}
{{- if .Values.auth.existingSecret -}}
{{- .Values.auth.existingSecret -}}
{{- else -}}
{{- printf "%s-aigateway-core-db" .Release.Name -}}
{{- end -}}
{{- end -}}

{{- define "higress-core-mysql.image.repository" -}}
{{- .Values.mysql.repository -}}
{{- end -}}

{{- define "higress-core-mysql.image.tag" -}}
{{- .Values.mysql.tag -}}
{{- end -}}

{{- define "higress-core-mysql.image.pullPolicy" -}}
{{- .Values.mysql.pullPolicy -}}
{{- end -}}

{{- define "higress-core-mysql.service.port" -}}
{{- .Values.mysql.port -}}
{{- end -}}
