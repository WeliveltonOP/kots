package template

import (
	"text/template"

	"github.com/replicatedhq/kots/kotskinds/apis/kots/v1beta1"
)

type VersionInfo struct {
	Sequence     int64  // the installation sequence. Always 0 when being freshly installed, etc
	Cursor       string // the upstream version cursor - integers for kots apps, may be semvers for helm charts
	ChannelName  string // the name of the channel that the current version was from
	VersionLabel string // a pretty version label if provided
	IsRequired   bool   // is the version/release required during upgrades or can it be skipped
	ReleaseNotes string // the release notes for the given version
	IsAirgap     bool   // is this an airgap app
}

type versionCtx struct {
	info *VersionInfo
}

func newVersionCtx(info *VersionInfo) versionCtx {
	return versionCtx{info: info}
}

func VersionInfoFromInstallation(sequence int64, isAirgap bool, spec v1beta1.InstallationSpec) VersionInfo {
	return VersionInfo{
		Sequence:     sequence,
		Cursor:       spec.UpdateCursor,
		ChannelName:  spec.ChannelName,
		VersionLabel: spec.VersionLabel,
		IsRequired:   spec.IsRequired,
		ReleaseNotes: spec.ReleaseNotes,
		IsAirgap:     isAirgap,
	}
}

// FuncMap represents the available functions in the versionCtx.
func (ctx versionCtx) FuncMap() template.FuncMap {
	return template.FuncMap{
		"Sequence":     ctx.sequence,
		"Cursor":       ctx.cursor,
		"ChannelName":  ctx.channelName,
		"VersionLabel": ctx.versionLabel,
		"IsRequired":   ctx.isRequired,
		"ReleaseNotes": ctx.releaseNotes,
		"IsAirgap":     ctx.isAirgap,
	}
}

func (ctx versionCtx) sequence() int64 {
	if ctx.info == nil {
		return -1
	}
	return ctx.info.Sequence
}

func (ctx versionCtx) cursor() string {
	if ctx.info == nil {
		return ""
	}
	return ctx.info.Cursor
}

func (ctx versionCtx) channelName() string {
	if ctx.info == nil {
		return ""
	}
	return ctx.info.ChannelName
}

func (ctx versionCtx) versionLabel() string {
	if ctx.info == nil {
		return ""
	}
	return ctx.info.VersionLabel
}

func (ctx versionCtx) isRequired() bool {
	if ctx.info == nil {
		return false
	}
	return ctx.info.IsRequired
}

func (ctx versionCtx) releaseNotes() string {
	if ctx.info == nil {
		return ""
	}
	return ctx.info.ReleaseNotes
}

func (ctx versionCtx) isAirgap() bool {
	if ctx.info == nil {
		return false
	}
	return ctx.info.IsAirgap
}
