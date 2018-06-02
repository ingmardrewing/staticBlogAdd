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
	bda.data = new(abstractData)

	imgFilename := bda.findImageFileInAddDir()
	imgPath := path.Join(addDir, imgFilename)
	bda.im = NewImageManager(bucket, imgPath)

	return bda
}

type abstractData struct {
	id            int
	htmlFilename  string
	imageFileName string
	title         string
	titlePlain    string
	microThumbUrl string
	thumbUrl      string
	imgUrl        string
	mdContent     string
	excerpt       string
	tags          []string
	url           string
	disqId        string
	content       string
	date          string
	category      string
}

type BlogDataAbstractor struct {
	data           *abstractData
	domain         string
	addDir         string
	postsDir       string
	defaultExcerpt string
	im             ImgManager
	dto            *staticIntf.PageDto
}

func (b *BlogDataAbstractor) ExtractData() {
	b.data.htmlFilename = "index.html"
	b.data.imageFileName = b.findImageFileInAddDir()

	title, titlePlain := b.inferBlogTitleFromFilename(b.data.imageFileName)
	b.data.title = title
	b.data.titlePlain = titlePlain

	microThumbUrl, thumbUrl, imgUrl, imgHtml := b.prepareImages()
	b.data.microThumbUrl = microThumbUrl
	b.data.thumbUrl = thumbUrl
	b.data.imgUrl = imgUrl

	mdContent, excerpt, tags := b.readMdData()
	b.data.mdContent = mdContent
	b.data.excerpt = excerpt
	b.data.tags = tags
	b.data.content = imgHtml + mdContent

	b.data.url = b.generateUrl(titlePlain)
	b.data.id = b.getId()
	b.data.disqId = b.generateDisqusId(b.data.id, titlePlain)
	b.data.date = staticUtil.GetDate()
	b.data.category = "blog post"
}

func (b *BlogDataAbstractor) GeneratePostDto() staticIntf.PageDto {
	return staticPersistence.NewFilledDto(
		b.data.id,
		b.data.title,
		b.data.titlePlain,
		b.data.thumbUrl,
		b.data.imgUrl,
		b.data.excerpt,
		b.data.disqId,
		b.data.date,
		b.data.content,
		b.data.url,
		b.domain,
		"",
		"",
		b.data.htmlFilename,
		"",
		b.data.category,
		b.data.microThumbUrl)
}

func (b *BlogDataAbstractor) GetTags() []string {
	return b.data.tags
}

func (b *BlogDataAbstractor) generateDisqusId(id int, titlePlain string) string {
	return fmt.Sprintf("%d %s%s", 1000000+id, b.domain, staticUtil.GenerateDatePath()+titlePlain)
}

func (b *BlogDataAbstractor) generateUrl(titlePlain string) string {
	return b.domain + staticUtil.GenerateDatePath() + titlePlain + "/"
}

func (b *BlogDataAbstractor) getId() int {
	postJsons := fs.ReadDirEntries(b.postsDir, false)
	if len(postJsons) == 0 {
		return 0
	}
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

func (b *BlogDataAbstractor) prepareImages() (string, string, string, string) {
	b.im.AddImageSize(190)
	b.im.AddImageSize(390)
	b.im.AddImageSize(800)
	b.im.PrepareImages()
	b.im.UploadImages()

	imgUrls := b.im.GetImageUrls()
	tpl := `<a href=\"%s\"><img src=\"%s\" width=\"800\"></a>`
	imgHtml := fmt.Sprintf(tpl, imgUrls[3], imgUrls[2])
	return imgUrls[0], imgUrls[1], imgUrls[2], imgHtml
}

func (b *BlogDataAbstractor) generateExcerpt(text string) string {
	text = b.stripLinksAndImages(text)
	if len(text) > 155 {
		txt := fmt.Sprintf("%.155s ...", text)
		return b.stripQuotes(txt)
	} else if len(text) == 0 {
		return b.defaultExcerpt
	}
	txt := strings.TrimSuffix(text, "\n")
	return b.stripQuotes(txt)
}

func (b *BlogDataAbstractor) generateHtmlFromMarkdown(input string) string {
	bytes := []byte(input)
	htmlBytes := blackfriday.Run(bytes, blackfriday.WithNoExtensions())
	htmlString := string(htmlBytes)
	trimmed := strings.TrimSuffix(htmlString, "\n")
	escaped := b.stripQuotes(trimmed)
	return strings.Replace(escaped, "\n", " ", -1)
}

// extracts social media hashtags from the given input
// and returns them as a slice of strings without the leading #
func (b *BlogDataAbstractor) extractTags(input string) []string {
	rx := regexp.MustCompile(`#[A-Za-zäüößÄÜÖ]+\b`)
	matches := rx.FindAllString(input, -1)
	resultSet := []string{}
	for _, m := range matches {
		resultSet = append(resultSet, strings.TrimPrefix(m, "#"))
	}
	return resultSet
}

func (b *BlogDataAbstractor) stripQuotes(txt string) string {
	txt = strings.Replace(txt, `'`, `’`, -1)
	return strings.Replace(txt, `"`, `\"`, -1)
}

func (b *BlogDataAbstractor) readMdData() (string, string, []string) {
	pathToMdFile := b.findMdFileInAddDir()
	if len(pathToMdFile) > 0 {
		mdData := fs.ReadFileAsString(pathToMdFile)
		excerpt := b.generateExcerpt(mdData)
		content := b.generateHtmlFromMarkdown(mdData)
		tags := b.extractTags(mdData)
		return content, excerpt, tags
	}
	return "", b.defaultExcerpt, []string{}
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

func (b *BlogDataAbstractor) inferBlogTitleFromFilename(filename string) (string, string) {
	fname := strings.TrimSuffix(filename, filepath.Ext(filename))
	return b.inferBlogTitle(fname), b.inferBlogTitlePlain(fname)
}

func (b *BlogDataAbstractor) inferBlogTitle(filename string) string {
	sepBySpecChars := splitAtSpecialChars(filename)
	parts := []string{}
	for _, s := range sepBySpecChars {
		parts = append(parts, splitCamelCaseAndNumbers(s)...)
	}

	spaceSeparated := strings.Join(parts, " ")
	return strings.Title(spaceSeparated)
}

func splitCamelCaseAndNumbers(whole string) []string {
	rx := regexp.MustCompile("([0-9]+|[A-ZÄÜÖ]?[a-zäüöß]*)")
	return rx.FindAllString(whole, -1)
}

func splitAtSpecialChars(whole string) []string {
	rx := regexp.MustCompile("[^-_ ,.]*")
	return rx.FindAllString(whole, -1)
}

func (b *BlogDataAbstractor) findMdFileInAddDir() string {
	mds := fs.ReadDirEntriesEndingWith(b.addDir, "md", "txt")
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
