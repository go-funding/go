package chromeoutputworker

import (
	"fuk-funding/go/utils"
	"fuk-funding/go/utils/ufiles"
	"os"
	"strings"
)

const DefaultTagsFileName = ".tags$$"
const DefaultTagsSeparator = "\n"

type Options struct {
	BaseDirAbsolute string
	TagsFileName    string
	TagsSeparator   string
}

type WebsiteTags struct {
	TagsFilePath string
	Separator    string
	tags         []string
}

func (wt WebsiteTags) AddTag(tag string) (err error) {
	return ufiles.AppendToFile(wt.TagsFilePath, tag+wt.Separator)
}

type Worker struct {
	options Options
}

func (w *Worker) GetTags(host string) (WebsiteTags, error) {
	tagsFilePath := w.getTagsFilePath(host)
	tags, err := os.ReadFile(tagsFilePath)
	if err != nil {
		return WebsiteTags{}, err
	}

	tagsList := strings.Split(string(tags), w.options.TagsSeparator)
	return WebsiteTags{
		TagsFilePath: tagsFilePath,
		Separator:    w.options.TagsSeparator,
		tags:         tagsList,
	}, nil
}

func (w *Worker) getTagsFilePath(host string) string {
	return utils.HostDirname(host)
}

func New(options Options) *Worker {
	if options.TagsFileName == "" {
		options.TagsFileName = DefaultTagsFileName
	}

	if options.TagsSeparator == "" {
		options.TagsSeparator = DefaultTagsSeparator
	}

	return &Worker{options}
}
