{{- if .Values.rbac.create }}
{{- if .Values.rbac.serviceAccount.create }}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ .Values.rbac.serviceAccount.name | quote }}
  labels:
    app: {{ template "fullname" . }}
    chart: "{{ .Chart.Name }}-{{ .Chart.Version }}"
    release: "{{ .Release.Name }}"
    heritage: "{{ .Release.Service }}"
  annotations:
    {{- range $k, $v := .Values.rbac.serviceAccount.annotations }}
    {{ $k | quote }}: {{ $v | quote }}
    {{- end }}
{{- end }}{{/* if .Values.rbac.serviceAccount.create */}}

{{- if .Values.rbac.namespaced.enabled }}
{{/* Checks for when the rbac is namespaced. */}}
{{- if .Values.defaultNamespace }}
{{- if not has .Values.defaultNamespace .Values.rbac.namespaced.whitelist }}
{{- fail "The default namespace is not whitelisted in rbac.namespaced.whitelist" }}
{{- end }}{{/* if not has .Values.defaultNamespace .Values.rbac.namespaced.whitelist */}}
{{- end }}{{/* if .Values.defaultNamespace */}}

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: {{ printf "service-broker-%s" .Release.Name | quote }}
  labels:
    app: {{ template "fullname" . }}
    chart: "{{ .Chart.Name }}-{{ .Chart.Version }}"
    release: "{{ .Release.Name }}"
    heritage: "{{ .Release.Service }}"
rules:
- apiGroups: [""]
  resources:
  - namespaces
  verbs:
  - get
  - list
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{ printf "service-broker-%s" .Release.Name | quote }}
  labels:
    app: {{ template "fullname" . }}
    chart: "{{ .Chart.Name }}-{{ .Chart.Version }}"
    release: "{{ .Release.Name }}"
    heritage: "{{ .Release.Service }}"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: {{ printf "service-broker-%s" .Release.Name | quote }}
subjects:
- kind: ServiceAccount
  name: {{ .Values.rbac.serviceAccount.name | quote }}
  namespace: {{ .Release.Namespace | quote }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: service-broker
  labels:
    app: {{ template "fullname" . }}
    chart: "{{ .Chart.Name }}-{{ .Chart.Version }}"
    release: "{{ .Release.Name }}"
    heritage: "{{ .Release.Service }}"
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: service-broker
  labels:
    app: {{ template "fullname" . }}
    chart: "{{ .Chart.Name }}-{{ .Chart.Version }}"
    release: "{{ .Release.Name }}"
    heritage: "{{ .Release.Service }}"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: service-broker
subjects:
- kind: ServiceAccount
  name: {{ .Values.rbac.serviceAccount.name | quote }}
  namespace: {{ .Release.Namespace | quote }}
{{- range $namespace := .Values.rbac.namespaced.whitelist }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: service-broker
  namespace: {{ $namespace | quote }}
  labels:
    app: {{ template "fullname" . }}
    chart: "{{ .Chart.Name }}-{{ .Chart.Version }}"
    release: "{{ .Release.Name }}"
    heritage: "{{ .Release.Service }}"
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: service-broker
  namespace: {{ $namespace | quote }}
  labels:
    app: {{ template "fullname" . }}
    chart: "{{ .Chart.Name }}-{{ .Chart.Version }}"
    release: "{{ .Release.Name }}"
    heritage: "{{ .Release.Service }}"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: service-broker
subjects:
- kind: ServiceAccount
  name: {{ $.Values.rbac.serviceAccount.name | quote }}
  namespace: {{ $.Release.Namespace | quote }}
{{- end }}{{/* range $namespace := .Values.rbac.namespaced.whitelist */}}
{{- else }}{{/* if .Values.rbac.namespaced.enabled */}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: {{ printf "service-broker-%s" .Release.Name | quote }}
  labels:
    app: {{ template "fullname" . }}
    chart: "{{ .Chart.Name }}-{{ .Chart.Version }}"
    release: "{{ .Release.Name }}"
    heritage: "{{ .Release.Service }}"
rules:
- apiGroups: [""]
  resources:
  - configmaps
  - endpoints
  - limitranges
  - persistentvolumeclaims
  - pods
  - podtemplates
  - replicationcontrollers
  - resourcequotas
  - secrets
  - serviceaccounts
  - services
  verbs: ["*"]
- apiGroups: [""]
  resources:
  - namespaces
  verbs:
  - get
  - list
- apiGroups:
  - apps
  - autoscaling
  - batch
  - networking.k8s.io
  resources: ["*"]
  verbs: ["*"]
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - rolebindings
  - roles
  verbs: ["*"]
- apiGroups:
  - policy
  resources:
  - poddisruptionbudgets
  verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: service-broker
  labels:
    app: {{ template "fullname" . }}
    chart: "{{ .Chart.Name }}-{{ .Chart.Version }}"
    release: "{{ .Release.Name }}"
    heritage: "{{ .Release.Service }}"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: {{ printf "service-broker-%s" .Release.Name | quote }}
subjects:
- kind: ServiceAccount
  name: {{ .Values.rbac.serviceAccount.name | quote }}
  namespace: {{ .Release.Namespace | quote }}
{{- end }}{{/* if .Values.rbac.namespaced.enabled */}}

{{- end }}{{/* if .Values.rbac.create */}}

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ template "fullname" . }}-client
  labels:
    app: {{ template "fullname" . }}
    chart: "{{ .Chart.Name }}--{{ .Chart.Version }}"
    release: "{{ .Release.Name }}"
    heritage: "{{ .Release.Service }}"
---
# Cluster role to grant the client service account the rights
# to call the /v2/* URLs that the broker serves
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: access-{{ template "fullname" . }}
  labels:
    app: {{ template "fullname" . }}
    chart: "{{ .Chart.Name }}--{{ .Chart.Version }}"
    release: "{{ .Release.Name }}"
    heritage: "{{ .Release.Service }}"
rules:
- nonResourceURLs: ["/v2", "/v2/*"]
  verbs: ["GET", "POST", "PUT", "PATCH", "DELETE"]

---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: {{ template "fullname" . }}-client
  labels:
    app: {{ template "fullname" . }}
    chart: "{{ .Chart.Name }}--{{ .Chart.Version }}"
    release: "{{ .Release.Name }}"
    heritage: "{{ .Release.Service }}"
subjects:
  - kind: ServiceAccount
    name: {{ template "fullname" . }}-client
    namespace: {{ .Release.Name }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: access-{{ template "fullname" . }}

---
# This secret needs to be a post install hook because otherwise it is skipped
# This causes the service catalog's cluster serverice broker to be unable to
# contact the broker.
apiVersion: v1
kind: Secret
metadata:
  name: {{ template "fullname" . }}
  annotations:
    kubernetes.io/service-account.name: {{ template "fullname" . }}-client
    "helm.sh/hook": post-install
    "helm.sh/hook-weight": "-5"
type: kubernetes.io/service-account-token