package staticBlogAdd

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ingmardrewing/fs"
	"github.com/ingmardrewing/staticIntf"
	"github.com/ingmardrewing/staticPersistence"
	"github.com/ingmardrewing/staticUtil"

	"gopkg.in/russross/blackfriday.v2"
)

func NewBlogDataAbstractor(bucket, addDir, postsDir, defaultExcerpt, domain string) *BlogDataAbstractor {
	bda := new(BlogDataAbstractor)
	bda.addDir = addDir
	bda.postsDir = postsDir
	bda.defaultExcerpt = defaultExcerpt
	bda.domain = domain

	imgFilename := bda.findImageFileInAddDir()
	imgPath := path.Join(addDir, imgFilename)
	bda.im = NewImageManager(bucket, imgPath)

	return bda
}

type BlogDataAbstractor struct {
	domain         string
	addDir         string
	postsDir       string
	defaultExcerpt string
	im             ImgManager
}

func (b *BlogDataAbstractor) GeneratePostDto() staticIntf.PageDto {
	htmlFilename := "index.html"
	title, titlePlain := b.inferBlogTitleFromFilename()
	thumbUrl, imgUrl, imgHtml := b.prepareImages()
	mdContent, excerpt := b.readMdData()
	url := b.generateUrl(titlePlain)
	id := b.getId()
	disqId := b.generateDisqusId(id, titlePlain)
	content := imgHtml + mdContent
	date := staticUtil.GetDate()

	return staticPersistence.NewFilledDto(
		0,
		title,
		titlePlain,
		thumbUrl,
		imgUrl,
		content,
		excerpt,
		disqId,
		date,
		url,
		b.domain,
		"",
		"",
		htmlFilename,
		"")
}

func (b *BlogDataAbstractor) generateDisqusId(id int, titlePlain string) string {
	return fmt.Sprintf("%d %s%s", 1000000+id, b.domain, staticUtil.GenerateDatePath()+titlePlain)
}

func (b *BlogDataAbstractor) generateUrl(titlePlain string) string {
	return b.domain + staticUtil.GenerateDatePath() + titlePlain + "/"
}

func (b *BlogDataAbstractor) getId() int {
	postJsons := fs.ReadDirEntries(b.postsDir, false)
	sort.Strings(postJsons)
	lastFile := postJsons[len(postJsons)-1]
	rx := regexp.MustCompile("(\\d+)")
	m := rx.FindStringSubmatch(lastFile)
	i, _ := strconv.Atoi(m[1])
	i++
	return i
}

func (b *BlogDataAbstractor) stripLinksAndImages(text string) string {
	rx := regexp.MustCompile(`\[.*\]\(.*\)`)
	return rx.ReplaceAllString(text, "")
}

func (b *BlogDataAbstractor) prepareImages() (string, string, string) {
	b.im.AddImageSize(390)
	b.im.AddImageSize(800)
	b.im.PrepareImages()
	b.im.UploadImages()

	imgUrls := b.im.GetImageUrls()
	tpl := `<a href=\"%s\"><img src=\"%s\" width=\"800\"></a>`
	imgHtml := fmt.Sprintf(tpl, imgUrls[2], imgUrls[1])
	return imgUrls[0], imgUrls[1], imgHtml
}

func (b *BlogDataAbstractor) generateExcerpt(text string) string {
	text = b.stripLinksAndImages(text)
	if len(text) > 155 {
		return fmt.Sprintf("%.155s ...", text)
	} else if len(text) == 0 {
		return b.defaultExcerpt
	}
	return strings.TrimSuffix(text, "\n")
}

func (b *BlogDataAbstractor) generateHtmlFromMarkdown(input string) string {
	bytes := []byte(input)
	htmlBytes := blackfriday.Run(bytes, blackfriday.WithNoExtensions())
	htmlString := string(htmlBytes)
	trimmed := strings.TrimSuffix(htmlString, "\n")
	escaped := strings.Replace(trimmed, `"`, `\"`, -1)
	return strings.Replace(escaped, "\n", " ", -1)
}

func (b *BlogDataAbstractor) readMdData() (string, string) {
	pathToMdFile := b.findMdFileInAddDir()
	if len(pathToMdFile) > 0 {
		mdData := fs.ReadFileAsString(pathToMdFile)
		excerpt := b.generateExcerpt(mdData)
		content := b.generateHtmlFromMarkdown(mdData)
		return content, excerpt
	}
	return "", b.defaultExcerpt
}

func (b *BlogDataAbstractor) findImageFileInAddDir() string {
	imgs := fs.ReadDirEntriesEndingWith(b.addDir, "png", "jpg")
	for _, i := range imgs {
		if !strings.Contains(i, "-w") {
			return i
		}
	}
	return ""
}

func (b *BlogDataAbstractor) inferBlogTitleFromFilename() (string, string) {
	filename := b.findImageFileInAddDir()
	fname := strings.TrimSuffix(filename, filepath.Ext(filename))
	return b.inferBlogTitle(fname), b.inferBlogTitlePlain(fname)
}

func (b *BlogDataAbstractor) inferBlogTitle(filename string) string {
	rx := regexp.MustCompile("(^[a-zäüöß]+)|([A-ZÄÜÖ][a-zäüöß,]*)|([0-9,]+)")
	parts := rx.FindAllString(filename, -1)
	spaceSeparated := strings.Join(parts, " ")
	return strings.Title(spaceSeparated)
}

func (b *BlogDataAbstractor) findMdFileInAddDir() string {
	mds := fs.ReadDirEntriesEndingWith(b.addDir, "md")
	for _, md := range mds {
		return path.Join(b.addDir, md)
	}
	return ""
}

func (b *BlogDataAbstractor) inferBlogTitlePlain(filename string) string {
	rx := regexp.MustCompile("(^[a-z]+)|([A-Z][a-z]*)|([0-9]+)")
	parts := rx.FindAllString(filename, -1)
	dashSeparated := strings.Join(parts, "-")
	return strings.ToLower(dashSeparated)
}
