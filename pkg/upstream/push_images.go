package upstream

import (
	"io"
	"time"

	"github.com/pkg/errors"
	kotsv1beta1 "github.com/replicatedhq/kots/kotskinds/apis/kots/v1beta1"
	"github.com/replicatedhq/kots/pkg/docker/registry"
	"github.com/replicatedhq/kots/pkg/image"
	"github.com/replicatedhq/kots/pkg/kotsadm"
	kotsadmtypes "github.com/replicatedhq/kots/pkg/kotsadm/types"
	"github.com/replicatedhq/kots/pkg/logger"
	"github.com/replicatedhq/kots/pkg/upstream/types"
	kustomizetypes "sigs.k8s.io/kustomize/api/types"
)

type ProcessUpstreamImagesOptions struct {
	RootDir             string
	ImagesDir           string
	AirgapBundle        string
	CreateAppDir        bool
	PushImages          bool
	UseKnownImages      bool
	KnownImages         []kustomizetypes.Image
	Log                 *logger.CLILogger
	ReplicatedRegistry  registry.RegistryOptions
	ReportWriter        io.Writer
	DestinationRegistry registry.RegistryOptions
}

type ProcessUpstreamImageResult struct {
	KustomizeImages []kustomizetypes.Image
	KnownImages     []kotsv1beta1.InstallationImage
}

func ProcessUpstreamImages(u *types.Upstream, options ProcessUpstreamImagesOptions) (*ProcessUpstreamImageResult, error) {
	pushOpts := kotsadmtypes.PushImagesOptions{
		Registry:       options.DestinationRegistry,
		Log:            options.Log,
		ProgressWriter: options.ReportWriter,
		LogForUI:       true,
	}

	var foundImages []kustomizetypes.Image
	if options.UseKnownImages {
		foundImages = options.KnownImages
	} else {
		if !options.PushImages {
			if options.AirgapBundle != "" {
				images, err := kotsadm.GetImagesFromBundle(options.AirgapBundle, pushOpts)
				if err != nil {
					return nil, errors.Wrap(err, "failed to push images")
				}
				foundImages = images
			} else {
				// TODO: Implement GetImagesFromPath
				return nil, errors.New("GetImagesFromPath is not implemented")
			}
		} else {
			if options.AirgapBundle != "" {
				images, err := kotsadm.TagAndPushAppImagesFromBundle(options.AirgapBundle, pushOpts)
				if err != nil {
					return nil, errors.Wrap(err, "failed to push images from bundle")
				}
				foundImages = images
			} else {
				images, err := kotsadm.TagAndPushAppImagesFromPath(options.ImagesDir, pushOpts)
				if err != nil {
					return nil, errors.Wrap(err, "failed to push images from dir")
				}
				foundImages = images
			}
		}
	}

	withAltNames := make([]kustomizetypes.Image, 0)
	for _, i := range foundImages {
		withAltNames = append(withAltNames, image.BuildImageAltNames(i)...)
	}

	result := &ProcessUpstreamImageResult{
		KustomizeImages: withAltNames,
		// This list is slightly different from the list we get from app specs because of alternative names,
		// but it still works because after rewriting image names with private registry, the lists become the same.
		KnownImages: makeInstallationImages(withAltNames),
	}
	return result, nil
}

type ProgressReport struct {
	// set to "progressReport"
	Type string `json:"type"`
	// the same progress text that used to be sent in unstructured message
	CompatibilityMessage string `json:"compatibilityMessage"`
	// all images found in archive
	Images []ProgressImage `json:"images"`
}

type ProgressImage struct {
	// image name and tag, "nginx:latest"
	DisplayName string `json:"displayName"`
	// image upload status: queued, uploading, uploaded, failed
	Status string `json:"status"`
	// error string set when status is failed
	Error string `json:"error"`
	// amount currently uploaded (currently number of layers)
	Current int64 `json:"current"`
	// total amount that needs to be uploaded (currently number of layers)
	Total int64 `json:"total"`
	// time when image started uploading
	StartTime time.Time `json:"startTime"`
	// time when image finished uploading
	EndTime time.Time `json:"endTime"`
}

func makeInstallationImages(images []kustomizetypes.Image) []kotsv1beta1.InstallationImage {
	result := make([]kotsv1beta1.InstallationImage, 0)
	for _, image := range images {
		result = append(result, kotsv1beta1.InstallationImage{
			Image:     image.Name,
			IsPrivate: true,
		})
	}
	return result
}
