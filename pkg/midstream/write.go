package midstream

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	kotsv1beta1 "github.com/replicatedhq/kots/kotskinds/apis/kots/v1beta1"
	"github.com/replicatedhq/kots/pkg/disasterrecovery"
	"github.com/replicatedhq/kots/pkg/k8sutil"
	"github.com/replicatedhq/kots/pkg/template"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	kustomizetypes "sigs.k8s.io/kustomize/api/types"
	k8syaml "sigs.k8s.io/yaml"
)

const (
	secretFilename                           = "secret.yaml"
	patchesFilename                          = "pullsecrets.yaml"
	disasterRecoveryLabelTransformerFileName = "backup-label-transformer.yaml"
)

type WriteOptions struct {
	MidstreamDir       string
	BaseDir            string
	AppSlug            string
	IsGitOps           bool
	IsOpenShift        bool
	Builder            template.Builder
	HTTPProxyEnvValue  string
	HTTPSProxyEnvValue string
	NoProxyEnvValue    string
	UseHelmInstall     map[string]bool
	NewHelmCharts      []*kotsv1beta1.HelmChart
}

func (m *Midstream) WriteMidstream(options WriteOptions) error {
	if err := os.MkdirAll(options.MidstreamDir, 0744); err != nil {
		return errors.Wrap(err, "failed to mkdir")
	}

	existingKustomization, err := m.getExistingKustomization(options)
	if err != nil {
		return errors.Wrap(err, "get existing kustomization")
	}

	secretFilename, err := m.writePullSecret(options)
	if err != nil {
		return errors.Wrap(err, "failed to write secret")
	}

	if secretFilename != "" {
		m.Kustomization.Resources = append(m.Kustomization.Resources, secretFilename)
	}

	identityBase, err := m.writeIdentityService(context.TODO(), options)
	if err != nil {
		return errors.Wrap(err, "failed to write identity service")
	}

	if identityBase != "" {
		m.Kustomization.Resources = append(m.Kustomization.Resources, identityBase)
	}

	if err := m.writeObjectsWithPullSecret(options); err != nil {
		return errors.Wrap(err, "failed to write patches")
	}

	// transformers
	drLabelTransformerFilename, err := m.writeDisasterRecoveryLabelTransformer(options)
	if err != nil {
		return errors.Wrap(err, "failed to write disaster recovery label transformer")
	}
	m.Kustomization.Transformers = append(m.Kustomization.Transformers, drLabelTransformerFilename)

	// annotations
	if m.Kustomization.CommonAnnotations == nil {
		m.Kustomization.CommonAnnotations = make(map[string]string)
	}
	m.Kustomization.CommonAnnotations["kots.io/app-slug"] = options.AppSlug

	if existingKustomization != nil {
		// Note that this function does nothing on the initial install
		// if the user is not presented with the config screen.
		m.mergeKustomization(options, *existingKustomization)
	}

	if err := m.writeKustomization(options); err != nil {
		return errors.Wrap(err, "failed to write kustomization")
	}

	return nil
}

func (m *Midstream) getExistingKustomization(options WriteOptions) (*kustomizetypes.Kustomization, error) {
	kustomizationFilename := filepath.Join(options.MidstreamDir, "kustomization.yaml")

	_, err := os.Stat(kustomizationFilename)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "stat existing kustomization")
	}

	k, err := k8sutil.ReadKustomizationFromFile(kustomizationFilename)
	if err != nil {
		return nil, errors.Wrap(err, "load existing kustomization")
	}
	return k, nil
}

func (m *Midstream) mergeKustomization(options WriteOptions, existing kustomizetypes.Kustomization) {
	existing.PatchesStrategicMerge = removeFromPatches(existing.PatchesStrategicMerge, patchesFilename)
	m.Kustomization.PatchesStrategicMerge = uniquePatches(existing.PatchesStrategicMerge, m.Kustomization.PatchesStrategicMerge)

	m.Kustomization.Resources = uniqueStrings(existing.Resources, m.Kustomization.Resources)

	m.Kustomization.Transformers = uniqueStrings(existing.Transformers, m.Kustomization.Transformers)

	// annotations
	if existing.CommonAnnotations == nil {
		existing.CommonAnnotations = make(map[string]string)
	}
	delete(existing.CommonAnnotations, "kots.io/app-sequence")
	m.Kustomization.CommonAnnotations = mergeMaps(m.Kustomization.CommonAnnotations, existing.CommonAnnotations)
}

func removeFromPatches(patches []kustomizetypes.PatchStrategicMerge, filename string) []kustomizetypes.PatchStrategicMerge {
	newPatches := []kustomizetypes.PatchStrategicMerge{}
	for _, patch := range patches {
		if string(patch) != filename {
			newPatches = append(newPatches, patch)
		}
	}
	return newPatches
}

func mergeMaps(new map[string]string, existing map[string]string) map[string]string {
	merged := existing
	if merged == nil {
		merged = make(map[string]string)
	}
	for key, value := range new {
		merged[key] = value
	}
	return merged
}

func (m *Midstream) writeKustomization(options WriteOptions) error {
	relativeBaseDir, err := filepath.Rel(options.MidstreamDir, options.BaseDir)
	if err != nil {
		return errors.Wrap(err, "failed to determine relative path for base from midstream")
	}

	kustomizationFilename := filepath.Join(options.MidstreamDir, "kustomization.yaml")

	m.Kustomization.Bases = []string{
		relativeBaseDir,
	}

	if err := k8sutil.WriteKustomizationToFile(*m.Kustomization, kustomizationFilename); err != nil {
		return errors.Wrap(err, "failed to write kustomization to file")
	}

	return nil
}

func (m *Midstream) writeDisasterRecoveryLabelTransformer(options WriteOptions) (string, error) {
	additionalLabels := map[string]string{
		"kots.io/app-slug": options.AppSlug,
	}
	drLabelTransformerYAML, err := disasterrecovery.GetLabelTransformerYAML(additionalLabels)
	if err != nil {
		return "", errors.Wrap(err, "failed to get disaster recovery label transformer yaml")
	}

	absFilename := filepath.Join(options.MidstreamDir, disasterRecoveryLabelTransformerFileName)

	if err := ioutil.WriteFile(absFilename, drLabelTransformerYAML, 0644); err != nil {
		return "", errors.Wrap(err, "failed to write disaster recovery label transformer yaml file")
	}

	return disasterRecoveryLabelTransformerFileName, nil
}

func (m *Midstream) writePullSecret(options WriteOptions) (string, error) {
	var secretBytes []byte
	if m.AppPullSecret != nil {
		b, err := k8syaml.Marshal(m.AppPullSecret)
		if err != nil {
			return "", errors.Wrap(err, "failed to marshal app pull secret")
		}
		secretBytes = b
	}

	if m.AdminConsolePullSecret != nil {
		if secretBytes != nil {
			secretBytes = append(secretBytes, []byte("\n---\n")...)
		}

		b, err := k8syaml.Marshal(m.AdminConsolePullSecret)
		if err != nil {
			return "", errors.Wrap(err, "failed to marshal kots pull secret")
		}
		secretBytes = append(secretBytes, b...)
	}

	if m.DockerHubPullSecret != nil {
		if secretBytes != nil {
			secretBytes = append(secretBytes, []byte("\n---\n")...)
		}

		b, err := k8syaml.Marshal(m.DockerHubPullSecret)
		if err != nil {
			return "", errors.Wrap(err, "failed to marshal kots pull secret")
		}
		secretBytes = append(secretBytes, b...)
	}

	if secretBytes == nil {
		return "", nil
	}

	absFilename := filepath.Join(options.MidstreamDir, secretFilename)
	if err := ioutil.WriteFile(absFilename, secretBytes, 0644); err != nil {
		return "", errors.Wrap(err, "failed to write pull secret file")
	}

	return secretFilename, nil
}

func (m *Midstream) writeObjectsWithPullSecret(options WriteOptions) error {
	filename := filepath.Join(options.MidstreamDir, patchesFilename)
	if len(m.DocForPatches) == 0 {
		err := os.Remove(filename)
		if err != nil && !os.IsNotExist(err) {
			return errors.Wrap(err, "failed to delete pull secret patches")
		}

		return nil
	}

	f, err := os.Create(filename)
	if err != nil {
		return errors.Wrap(err, "failed to create resources file")
	}
	defer f.Close()

	secrets := []*corev1.Secret{}
	if m.AppPullSecret != nil {
		secrets = append(secrets, m.AppPullSecret)
	}
	if m.DockerHubPullSecret != nil {
		secrets = append(secrets, m.DockerHubPullSecret)
	}

	for _, o := range m.DocForPatches {
		for _, secret := range secrets {
			withPullSecret := o.PatchWithPullSecret(secret)
			if withPullSecret == nil {
				continue
			}

			b, err := yaml.Marshal(withPullSecret)
			if err != nil {
				return errors.Wrap(err, "failed to marshal object")
			}

			if _, err := f.Write([]byte("---\n")); err != nil {
				return errors.Wrap(err, "failed to write doc separator")
			}
			if _, err := f.Write(b); err != nil {
				return errors.Wrap(err, "failed to write object")
			}
		}
	}

	m.Kustomization.PatchesStrategicMerge = append(m.Kustomization.PatchesStrategicMerge, patchesFilename)

	return nil
}

func EnsureDisasterRecoveryLabelTransformer(archiveDir string, additionalLabels map[string]string) error {
	labelTransformerExists := false

	dirPath := filepath.Join(archiveDir, "overlays", "midstream")

	// TODO (ch35027): this will not work with multiple kustomization files
	k, err := k8sutil.ReadKustomizationFromFile(filepath.Join(dirPath, "kustomization.yaml"))
	if err != nil {
		return errors.Wrap(err, "failed to read kustomization file from midstream")
	}

	for _, transformer := range k.Transformers {
		if transformer == disasterRecoveryLabelTransformerFileName {
			labelTransformerExists = true
			break
		}
	}

	if !labelTransformerExists {
		drLabelTransformerYAML, err := disasterrecovery.GetLabelTransformerYAML(additionalLabels)
		if err != nil {
			return errors.Wrap(err, "failed to get disaster recovery label transformer yaml")
		}

		absFilename := filepath.Join(dirPath, disasterRecoveryLabelTransformerFileName)

		if err := ioutil.WriteFile(absFilename, drLabelTransformerYAML, 0644); err != nil {
			return errors.Wrap(err, "failed to write disaster recovery label transformer yaml file")
		}

		k.Transformers = append(k.Transformers, disasterRecoveryLabelTransformerFileName)

		if err := k8sutil.WriteKustomizationToFile(*k, filepath.Join(dirPath, "kustomization.yaml")); err != nil {
			return errors.Wrap(err, "failed to write kustomization file to midstream")
		}
	}

	return nil
}
