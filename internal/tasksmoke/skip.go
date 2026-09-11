// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"runtime"
	"slices"
	"strings"

	"github.com/task-otter/store/internal/tasktest"
)

func destructiveSkipReason(name string) string {
	if isCleanSkip(name) || nameMatchesAny(name, extraSkipNames()) {
		return reasonDestructive
	}

	return emptyString
}

func dockerSkipReason(module, name string) string {
	if toolName(module) != toolDocker || !nameMatchesAny(name, dockerWorkNames()) {
		return emptyString
	}

	return reasonDocker
}

func dockerWorkNames() []string {
	return []string{
		nameBuild, nameImages, namePrune, namePruneAll, namePS, namePSAll, namePull, nameStopAll,
		nameVerify, nameVersion,
	}
}

func extraSkipNames() []string {
	return []string{
		nameUninstall, nameInstallUndo, nameCIFix, nameLintFix, nameUpgrade,
		namePush, namePushForce, namePull, nameFetch, nameClone, nameResetHard, nameResetSoft,
		nameTagDelete, nameTagPush, namePRCreate, namePROpen, nameReleaseCreate, nameRemoteRemove,
		nameStashDrop, nameCommitAmend, nameBranchDelete, nameVaultEncrypt, nameVaultDecrypt,
	}
}

func fuzzSkipReason(name string, vars map[string]string) string {
	if !nameMatches(name, nameFuzz) || hasNonEmptyVar(vars, nameCLIArgs) {
		return emptyString
	}

	return reasonFuzz
}

func ghAllowedNames() []string {
	return []string{
		nameAliasList, nameConfigList, nameHelp, nameInstall, nameVerify, nameVersion, nameWhich,
	}
}

func ghAndYAMLSkipReason(input *skipInput) string {
	reason := ghSkipReason(input.Module, input.Name)

	if reason != emptyString {
		return reason
	}

	return yamlAndNixSkipReason(input)
}

func ghSkipReason(module, name string) string {
	if toolName(module) != toolGH || slices.Contains(ghAllowedNames(), name) {
		return emptyString
	}

	return reasonGH
}

func hasNonEmptyVar(vars map[string]string, key string) bool {
	if vars == nil {
		return false
	}

	return strings.TrimSpace(vars[key]) != emptyString
}

func hasPrompt(prompt any) bool {
	if prompt == nil {
		return false
	}

	return promptIsPresent(prompt)
}

func interactiveOrRequiredSkip(input *skipInput) string {
	if input.Task != nil && input.Task.Interactive {
		return reasonInteractive
	}

	return requiredVarsSkipReason(taskRequiresVars(input.Task), smokeVars(input.Config))
}

func isCleanSkip(name string) bool {
	if name == nameCacheClean {
		return false
	}

	return nameMatches(name, nameClean)
}

func isFmtRewrite(name string) bool {
	if name == nameFmt {
		return true
	}

	return strings.HasSuffix(name, suffixFmt)
}

func nameMatches(name, rule string) bool {
	if name == rule {
		return true
	}

	return strings.HasSuffix(name, colonSeparator+rule)
}

func nameMatchesAny(name string, rules []string) bool {
	for i := range rules {
		if nameMatches(name, rules[i]) {
			return true
		}
	}

	return false
}

func nameSkipReason(name string) string {
	if isFmtRewrite(name) {
		return reasonFmt
	}

	if strings.HasSuffix(name, suffixWatch) {
		return reasonWatch
	}

	return destructiveSkipReason(name)
}

func nixSkipReason(module, name string) string {
	if module != toolNix || !nameMatches(name, nameInstall) {
		return emptyString
	}

	return reasonNixInstall
}

func platformSkipReason(input *skipInput) string {
	reason := wingetSkipReason(input.Module)

	if reason != emptyString {
		return reason
	}

	return unixOnlyAndGHSkipReason(input)
}

func policySkipReason(input *skipInput) string {
	reason := nameSkipReason(input.Name)

	if reason != emptyString {
		return reason
	}

	return scopedSkipReason(input)
}

func promptIsPresent(prompt any) bool {
	text, isString := prompt.(string)

	if isString {
		return strings.TrimSpace(text) != emptyString
	}

	return promptListPresent(prompt)
}

func promptListPresent(prompt any) bool {
	items, isList := prompt.([]any)

	if isList {
		return len(items) > emptyLength
	}

	return promptStringListPresent(prompt)
}

func promptSkipReason(task *tasktest.Task) string {
	if task == nil || !hasPrompt(task.Prompt) {
		return emptyString
	}

	return reasonPrompt
}

func promptStringListPresent(prompt any) bool {
	items, isList := prompt.([]string)

	if !isList {
		return true
	}

	return len(items) > emptyLength
}

func requiredVarsSkipReason(names []string, vars map[string]string) string {
	for i := range names {
		if !hasNonEmptyVar(vars, names[i]) {
			return reasonRequired
		}
	}

	return emptyString
}

func scopedSkipReason(input *skipInput) string {
	reason := dockerSkipReason(input.Module, input.Name)

	if reason != emptyString {
		return reason
	}

	return platformSkipReason(input)
}

func skipReason(input *skipInput) string {
	reason := taskFieldSkipReason(input)

	if reason != emptyString {
		return reason
	}

	return policySkipReason(input)
}

func smokeSkips(config *smokeConfig) []string {
	if config == nil {
		return nil
	}

	return config.Skip
}

func smokeVars(config *smokeConfig) map[string]string {
	if config == nil {
		return nil
	}

	return config.Vars
}

func taskFieldSkipReason(input *skipInput) string {
	reason := promptSkipReason(input.Task)

	if reason != emptyString {
		return reason
	}

	return interactiveOrRequiredSkip(input)
}

func taskRequiresVars(task *tasktest.Task) []string {
	if task == nil {
		return nil
	}

	return requiredVarNames(task.Requires)
}

func unixOnlyAndGHSkipReason(input *skipInput) string {
	reason := unixOnlySkipReason(input.Module)

	if reason != emptyString {
		return reason
	}

	reason = cargoSourceSkipReason(input.Module)

	if reason != emptyString {
		return reason
	}

	return ghAndYAMLSkipReason(input)
}

func unixOnlyModules() []string {
	return []string{toolAnsible, toolAnsibleLint, toolNix}
}

func unixOnlySkipOnOS(module, goos string) string {
	if goos != osWindows || !slices.Contains(unixOnlyModules(), toolName(module)) {
		return emptyString
	}

	return reasonUnixOnly
}

func unixOnlySkipReason(module string) string {
	return unixOnlySkipOnOS(module, runtime.GOOS)
}

// cargoSourceModules need a full Rust toolchain + cargo install on Windows
// (no winget package). That is too slow/flaky for GHA Windows smoke.
func cargoSourceModules() []string {
	return []string{toolAdrs}
}

func cargoSourceSkipOnOS(module, goos string) string {
	if goos != osWindows || !slices.Contains(cargoSourceModules(), toolName(module)) {
		return emptyString
	}

	return reasonCargoSource
}

func cargoSourceSkipReason(module string) string {
	return cargoSourceSkipOnOS(module, runtime.GOOS)
}

func wingetSkipOnOS(module, goos string) string {
	if toolName(module) != toolWinget || goos == osWindows {
		return emptyString
	}

	return reasonWinget
}

func wingetSkipReason(module string) string {
	return wingetSkipOnOS(module, runtime.GOOS)
}

func yamlAndNixSkipReason(input *skipInput) string {
	reason := nixSkipReason(input.Module, input.Name)

	if reason != emptyString {
		return reason
	}

	return yamlFuzzSkipReason(input)
}

func yamlFuzzSkipReason(input *skipInput) string {
	reason := yamlSkipReason(input.Name, smokeSkips(input.Config))

	if reason != emptyString {
		return reason
	}

	return fuzzSkipReason(input.Name, smokeVars(input.Config))
}

func yamlSkipReason(name string, skip []string) string {
	if slices.Contains(skip, name) {
		return reasonYAMLSkip
	}

	return emptyString
}
