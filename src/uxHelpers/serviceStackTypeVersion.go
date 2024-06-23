package uxHelpers

import (
	"context"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/options"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
)

type printServiceStackTypeVersionSelector struct {
	filters []func(entity.ServiceStackType, entity.ServiceStackTypeVersion) bool
}

func PrintServiceStackTypeVersionSelectorWithServiceStackTypeIdFilter(in entity.ServiceStackType) options.Option[printServiceStackTypeVersionSelector] {
	return func(p *printServiceStackTypeVersionSelector) {
		p.filters = append(p.filters, func(t entity.ServiceStackType, _ entity.ServiceStackTypeVersion) bool {
			return t.ID == in.ID
		})
	}
}

func PrintServiceStackTypeVersionSelector(
	ctx context.Context,
	uxBlocks uxBlock.UxBlocks,
	restApiClient *zeropsRestApiClient.Handler,
	opts ...options.Option[printServiceStackTypeVersionSelector],
) (entity.ServiceStackTypeVersion, error) {
	setup := options.ApplyOptions(opts...)

	list, err := repository.GetServiceStackTypes(ctx, restApiClient)
	if err != nil {
		return entity.ServiceStackTypeVersion{}, err
	}
	var versionList []entity.ServiceStackTypeVersion
	for _, typeItem := range list {
		for _, versionItem := range typeItem.Versions {
			var skipVersion bool
			for _, filter := range setup.filters {
				if !filter(typeItem, versionItem) {
					skipVersion = true
					break
				}
			}
			if skipVersion {
				continue
			}
			versionList = append(versionList, versionItem)
		}
	}

	return SelectOne(ctx, uxBlocks, versionList,
		SelectOneWithHeader[entity.ServiceStackTypeVersion]("ID", "Name"),
		SelectOneWithSelectLabel[entity.ServiceStackTypeVersion](i18n.T(i18n.ServiceStackTypeVersionSelectorPrompt)),
		SelectOneWithNotFound[entity.ServiceStackTypeVersion](i18n.T(i18n.ServiceStackTypeVersionSelectorOutOfRangeError)),
		SelectOneWithRow(func(in entity.ServiceStackTypeVersion) []string {
			return []string{string(in.ID), in.Name.String()}
		}),
	)
}

type selectOneConfig[T any] struct {
	Header      []string
	Row         func(T) []string
	SelectLabel string
	NotFound    string
}

func SelectOneWithSelectLabel[T any](label string) options.Option[selectOneConfig[T]] {
	return func(s *selectOneConfig[T]) {
		s.SelectLabel = label
	}
}

func SelectOneWithHeader[T any](columns ...string) options.Option[selectOneConfig[T]] {
	return func(s *selectOneConfig[T]) {
		s.Header = columns
	}
}

func SelectOneWithRow[T any](in func(T) []string) options.Option[selectOneConfig[T]] {
	return func(s *selectOneConfig[T]) {
		s.Row = in
	}
}

func SelectOneWithNotFound[T any](label string) options.Option[selectOneConfig[T]] {
	return func(s *selectOneConfig[T]) {
		s.NotFound = label
	}
}

func SelectOne[T any](ctx context.Context, uxBlocks uxBlock.UxBlocks, list []T, opts ...options.Option[selectOneConfig[T]]) (result T, _ error) {
	setup := options.ApplyOptions(opts...)

	if len(list) == 0 {
		uxBlocks.PrintWarning(styles.WarningLine(setup.NotFound))
		return result, errors.New(setup.NotFound)
	}
	header := (&uxBlock.TableRow{}).AddStringCells(setup.Header...)

	tableBody := &uxBlock.TableBody{}
	for _, listItem := range list {
		tableBody.AddStringsRow(setup.Row(listItem)...)
	}

	listIndex, err := uxBlocks.Select(
		ctx,
		tableBody,
		uxBlock.SelectLabel(setup.SelectLabel),
		uxBlock.SelectTableHeader(header),
	)

	if err != nil {
		return result, err
	}

	if len(listIndex) == 0 {
		return result, errors.New(setup.NotFound)
	}

	if listIndex[0] > len(list)-1 {
		return result, errors.New(setup.NotFound)
	}

	return list[listIndex[0]], nil
}
