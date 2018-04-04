package staticBlogAdd

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/ingmardrewing/fs"
	"github.com/ingmardrewing/staticPersistence"
	"github.com/ingmardrewing/staticUtil"
)

func TestMain(m *testing.M) {
	//setup()
	code := m.Run()
	tearDown()
	os.Exit(code)
}
func tearDown() {
	p := path.Join(getTestFileDirPath(), "testResources/src/posts/")
	fs.RemoveFile(p, "page342.json")
	//	fs.RemoveFile(p, "TestImage-w800.png")
}

func getTestFileDirPath() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}

func TestGenerateDatePath(t *testing.T) {
	actual := staticUtil.GenerateDatePath()
	now := time.Now()
	expected := fmt.Sprintf("%d/%d/%d/", now.Year(), now.Month(), now.Day())

	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func TestBlogDataAbstractor(t *testing.T) {
	addDir := getTestFileDirPath() + "/testResources/src/add/"
	postsDir := getTestFileDirPath() + "/testResources/src/posts/"
	dExcerpt := "A blog containing texts, drawings, graphic narratives/novels and (rarely) code snippets by Ingmar Drewing."

	bda := NewBlogDataAbstractor("drewingde", addDir, postsDir, dExcerpt, "https://drewing.de/blog/")
	bda.im = &imgManagerMock{}
	dto := bda.GeneratePostDto()

	actual := dto.Title()
	expected := "Test Image"

	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}

	actual = dto.Content()
	expected = `<a href=\"https://drewing.de/just/another/path/TestImage.png\"><img src=\"https://drewing.de/just/another/path/TestImage-w800.png\" width=\"800\"></a><p>This is a Test</p>`
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}

	actual = dto.Description()
	expected = `This is a Test`

	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}

	actual = dto.CreateDate()
	n := time.Now()
	expected = fmt.Sprintf("%d-%d-%d %d:%d:%d", n.Year(), n.Month(), n.Day(), n.Hour(), n.Minute(), n.Second())

	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}

	actualInt := dto.Id()
	expectedInt := 342

	if actualInt != expectedInt {
		t.Errorf("Expected %d, but got %d\n", expectedInt, actualInt)
	}
}

func TestWriteData(t *testing.T) {
	addDir := getTestFileDirPath() + "/testResources/src/add/"
	postsDir := getTestFileDirPath() + "/testResources/src/posts/"
	dExcerpt := "A blog containing texts, drawings, graphic narratives/novels and (rarely) code snippets by Ingmar Drewing."
	domain := "https://drewing.de/blog/"

	bda := NewBlogDataAbstractor("drewingde", addDir, postsDir, dExcerpt, domain)
	bda.im = &imgManagerMock{}
	dto := bda.GeneratePostDto()

	filename := fmt.Sprintf("page%d.json", dto.Id())

	staticPersistence.WritePageDtoToJson(dto, postsDir, filename)

	data := fs.ReadFileAsString(path.Join(postsDir, filename))
	tpl := `{
	"version":1,
	"thumbImg":"https://drewing.de/just/another/path/TestImage-w390.png",
	"postImg":"https://drewing.de/just/another/path/TestImage-w800.png",
	"filename":"index.html",
	"id":342,
	"date":"%s",
	"url":"%s",
	"title":"Test Image",
	"title_plain":"test-image",
	"excerpt":"This is a Test",
	"content":"<a href=\"https://drewing.de/just/another/path/TestImage.png\"><img src=\"https://drewing.de/just/another/path/TestImage-w800.png\" width=\"800\"></a><p>This is a Test</p>",
	"dsq_thread_id":"%s"
}`
	dp := staticUtil.GenerateDatePath()
	dsq := fmt.Sprintf("%d %s%s", 1000000+dto.Id(), domain, dp+dto.TitlePlain())
	url := fmt.Sprintf("https://drewing.de/blog/%stest-image/", dp)
	expected := fmt.Sprintf(tpl, staticUtil.GetDate(), url, dsq)

	if data != expected {
		t.Error("Expected", expected, "but got", data)
	}
}

type imgManagerMock struct{}

func (i *imgManagerMock) PrepareImages() {}
func (i *imgManagerMock) UploadImages()  {}
func (i *imgManagerMock) GetImageUrls() []string {
	return []string{
		"https://drewing.de/just/another/path/TestImage-w390.png",
		"https://drewing.de/just/another/path/TestImage-w800.png",
		"https://drewing.de/just/another/path/TestImage.png"}
}
func (i *imgManagerMock) AddImageSize(size int) string {
	return "TestImage-w" + strconv.Itoa(size) + ".png"
}
