package staticBlogAdd

import (
	"strconv"
	"strings"

	"github.com/ingmardrewing/aws"
	"github.com/ingmardrewing/fs"
	"github.com/ingmardrewing/img"
	"github.com/ingmardrewing/staticUtil"
)

var doUpload = true

func DoUpload(val bool) {
	doUpload = val
}

type ImgManager interface {
	PrepareImages()
	UploadImages()
	GetImageUrls() []string
	AddImageSize(size int) string
}

// Upload images to the given awsbucket using
// environmental data as required by the aws
// packages
func NewImageManager(awsbucket, sourceimagepath string) *ImageManager {
	im := new(ImageManager)
	im.awsbucket = awsbucket
	im.sourceimagepath = sourceimagepath
	return im
}

type ImageManager struct {
	sourceimagepath   string
	uploadimgagepaths []string
	awsimageurls      []string
	imagesizes        []int
	awsbucket         string
}

func (i *ImageManager) PrepareImages() {
	imgdir := fs.GetPathWithoutFilename(i.sourceimagepath)
	img := img.NewImgScaler(i.sourceimagepath, imgdir)
	paths := img.PrepareResizeTo(i.imagesizes...)
	i.uploadimgagepaths = append(paths, i.sourceimagepath)
	img.Resize()
}

func (i *ImageManager) UploadImages() {
	if !doUpload {
		return
	}
	for _, filepath := range i.uploadimgagepaths {
		filename := fs.GetFilenameFromPath(filepath)
		key := i.getS3Key(filename)
		url := aws.UploadFile(filepath, i.awsbucket, key)
		i.awsimageurls = append(i.awsimageurls, url)
	}
}

func (i *ImageManager) getS3Key(filename string) string {
	return "blog/" + staticUtil.GenerateDatePath() + filename
}

func (i *ImageManager) GetImageUrls() []string {
	if !doUpload {
		return i.uploadimgagepaths
	}
	return i.awsimageurls
}

func (i *ImageManager) AddImageSize(size int) string {
	i.imagesizes = append(i.imagesizes, size)
	return i.getFileNameFor(size)
}

func (i *ImageManager) getFileNameFor(w int) string {
	tag := "-w" + strconv.Itoa(w)

	sf := fs.GetFilenameFromPath(i.sourceimagepath)
	parts := strings.Split(sf, ".")
	n := strings.Join(parts[:len(parts)-1], "")
	return n + tag + "." + parts[len(parts)-1]
}
