package staticBlogAdd

import (
	"fmt"
	"os"
	"path"
	"reflect"
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
	pth := path.Join(getTestFileDirPath(), "testResources/src/posts/")
	filename := "page358.json"
	if exist, _ := fs.PathExists(path.Join(pth, filename)); exist == true {
		fs.RemoveFile(pth, filename)
	}
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
	bda := givenBlogDataAbstractor()
	bda.im = &imgManagerMock{}
	dto := bda.GeneratePostDto()

	actual := dto.Title()
	expected := "Test Image"

	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}

	actual = dto.Content()
	expected = `<a href=\"https://drewing.de/just/another/path/TestImage.png\"><img src=\"https://drewing.de/just/another/path/TestImage-w800.png\" width=\"800\"></a>`
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}

	actual = dto.Description()
	expected = "A blog containing texts, drawings, graphic narratives/novels and (rarely) code snippets by Ingmar Drewing."

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
	expectedInt := 358

	if actualInt != expectedInt {
		t.Errorf("Expected %d, but got %d\n", expectedInt, actualInt)
	}
}

func TestSplitAtSpecialChars(t *testing.T) {
	expected := []string{"a", "b", "c", "d"}
	actual := splitAtSpecialChars("a-b,c_d")

	if !reflect.DeepEqual(expected, actual) {
		t.Error("Expected", expected, "but got", actual)
	}
}

func TestSplitCamelCaseAndNumbers(t *testing.T) {
	expected := []string{"another", "Test", "4", "this"}
	actual := splitCamelCaseAndNumbers("anotherTest4this")

	if !reflect.DeepEqual(expected, actual) {
		t.Error("Expected", expected, "but got", actual)
	}
}

func TestInferBlogTitleFromFilename(t *testing.T) {
	bda := givenBlogDataAbstractor()

	filename2expected := map[string]string{
		"iPadTest.png":       "I Pad Test",
		"this-is-a-test.png": "This Is A Test",
		"test_image.png":     "Test Image",
		"even4me.png":        "Even 4 Me"}
	for filename, expected := range filename2expected {
		actual, _ := bda.inferBlogTitleFromFilename(filename)
		if actual != expected {
			t.Error("Expected", expected, "but got", actual)
		}
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
	"id":358,
	"date":"%s",
	"url":"%s",
	"title":"Test Image",
	"title_plain":"test-image",
	"excerpt":"A blog containing texts, drawings, graphic narratives/novels and (rarely) code snippets by Ingmar Drewing.",
	"content":"<a href=\"https://drewing.de/just/another/path/TestImage.png\"><img src=\"https://drewing.de/just/another/path/TestImage-w800.png\" width=\"800\"></a>",
	"dsq_thread_id":"%s",
	"thumbBase64":"%s",
	"category":"%s"
}`
	dp := staticUtil.GenerateDatePath()
	dsq := fmt.Sprintf("%d %s%s", 1000000+dto.Id(), domain, dp+dto.TitlePlain())
	url := fmt.Sprintf("https://drewing.de/blog/%stest-image/", dp)
	expected := fmt.Sprintf(tpl, staticUtil.GetDate(), url, dsq, "", "blog post")

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

func givenBlogDataAbstractor() *BlogDataAbstractor {
	addDir := getTestFileDirPath() + "/testResources/src/add/"
	postsDir := getTestFileDirPath() + "/testResources/src/posts/"
	dExcerpt := "A blog containing texts, drawings, graphic narratives/novels and (rarely) code snippets by Ingmar Drewing."

	return NewBlogDataAbstractor("drewingde",
		addDir,
		postsDir,
		dExcerpt,
		"https://drewing.de/blog/")
}
