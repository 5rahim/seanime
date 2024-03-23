package tvdb

//goland:noinspection GoSnakeCaseUsage
type ExtendedSeriesResponse_Season struct {
	ID                   int64    `json:"id,omitempty"`
	Image                string   `json:"image,omitempty"`
	ImageType            int64    `json:"imageType,omitempty"`
	LastUpdated          string   `json:"lastUpdated,omitempty"`
	Name                 string   `json:"name,omitempty"`
	NameTranslations     []string `json:"nameTranslations,omitempty"`
	Number               int64    `json:"number,omitempty"`
	OverviewTranslations []string `json:"overviewTranslations,omitempty"`
	Companies            struct {
		Studio []struct {
			ActiveDate string `json:"activeDate,omitempty"`
			Aliases    []struct {
				Language string `json:"language,omitempty"`
				Name     string `json:"name,omitempty"`
			} `json:"aliases,omitempty"`
			Country              string   `json:"country,omitempty"`
			ID                   int64    `json:"id,omitempty"`
			InactiveDate         string   `json:"inactiveDate,omitempty"`
			Name                 string   `json:"name,omitempty"`
			NameTranslations     []string `json:"nameTranslations,omitempty"`
			OverviewTranslations []string `json:"overviewTranslations,omitempty"`
			PrimaryCompanyType   int64    `json:"primaryCompanyType,omitempty"`
			Slug                 string   `json:"slug,omitempty"`
			ParentCompany        struct {
				ID       int64  `json:"id,omitempty"`
				Name     string `json:"name,omitempty"`
				Relation struct {
					ID       int64  `json:"id,omitempty"`
					TypeName string `json:"typeName,omitempty"`
				} `json:"relation,omitempty"`
			} `json:"parentCompany,omitempty"`
			TagOptions []struct {
				HelpText string `json:"helpText,omitempty"`
				ID       int64  `json:"id,omitempty"`
				Name     string `json:"name,omitempty"`
				Tag      int64  `json:"tag,omitempty"`
				TagName  string `json:"tagName,omitempty"`
			} `json:"tagOptions,omitempty"`
		} `json:"studio,omitempty"`
		Network []struct {
			ActiveDate string `json:"activeDate,omitempty"`
			Aliases    []struct {
				Language string `json:"language,omitempty"`
				Name     string `json:"name,omitempty"`
			} `json:"aliases,omitempty"`
			Country              string   `json:"country,omitempty"`
			ID                   int64    `json:"id,omitempty"`
			InactiveDate         string   `json:"inactiveDate,omitempty"`
			Name                 string   `json:"name,omitempty"`
			NameTranslations     []string `json:"nameTranslations,omitempty"`
			OverviewTranslations []string `json:"overviewTranslations,omitempty"`
			PrimaryCompanyType   int64    `json:"primaryCompanyType,omitempty"`
			Slug                 string   `json:"slug,omitempty"`
			ParentCompany        struct {
				ID       int64  `json:"id,omitempty"`
				Name     string `json:"name,omitempty"`
				Relation struct {
					ID       int64  `json:"id,omitempty"`
					TypeName string `json:"typeName,omitempty"`
				} `json:"relation,omitempty"`
			} `json:"parentCompany,omitempty"`
			TagOptions []struct {
				HelpText string `json:"helpText,omitempty"`
				ID       int64  `json:"id,omitempty"`
				Name     string `json:"name,omitempty"`
				Tag      int64  `json:"tag,omitempty"`
				TagName  string `json:"tagName,omitempty"`
			} `json:"tagOptions,omitempty"`
		} `json:"network,omitempty"`
		Production []struct {
			ActiveDate string `json:"activeDate,omitempty"`
			Aliases    []struct {
				Language string `json:"language,omitempty"`
				Name     string `json:"name,omitempty"`
			} `json:"aliases,omitempty"`
			Country              string   `json:"country,omitempty"`
			ID                   int64    `json:"id,omitempty"`
			InactiveDate         string   `json:"inactiveDate,omitempty"`
			Name                 string   `json:"name,omitempty"`
			NameTranslations     []string `json:"nameTranslations,omitempty"`
			OverviewTranslations []string `json:"overviewTranslations,omitempty"`
			PrimaryCompanyType   int64    `json:"primaryCompanyType,omitempty"`
			Slug                 string   `json:"slug,omitempty"`
			ParentCompany        struct {
				ID       int64  `json:"id,omitempty"`
				Name     string `json:"name,omitempty"`
				Relation struct {
					ID       int64  `json:"id,omitempty"`
					TypeName string `json:"typeName,omitempty"`
				} `json:"relation,omitempty"`
			} `json:"parentCompany,omitempty"`
			TagOptions []struct {
				HelpText string `json:"helpText,omitempty"`
				ID       int64  `json:"id,omitempty"`
				Name     string `json:"name,omitempty"`
				Tag      int64  `json:"tag,omitempty"`
				TagName  string `json:"tagName,omitempty"`
			} `json:"tagOptions,omitempty"`
		} `json:"production,omitempty"`
		Distributor []struct {
			ActiveDate string `json:"activeDate,omitempty"`
			Aliases    []struct {
				Language string `json:"language,omitempty"`
				Name     string `json:"name,omitempty"`
			} `json:"aliases,omitempty"`
			Country              string   `json:"country,omitempty"`
			ID                   int64    `json:"id,omitempty"`
			InactiveDate         string   `json:"inactiveDate,omitempty"`
			Name                 string   `json:"name,omitempty"`
			NameTranslations     []string `json:"nameTranslations,omitempty"`
			OverviewTranslations []string `json:"overviewTranslations,omitempty"`
			PrimaryCompanyType   int64    `json:"primaryCompanyType,omitempty"`
			Slug                 string   `json:"slug,omitempty"`
			ParentCompany        struct {
				ID       int64  `json:"id,omitempty"`
				Name     string `json:"name,omitempty"`
				Relation struct {
					ID       int64  `json:"id,omitempty"`
					TypeName string `json:"typeName,omitempty"`
				} `json:"relation,omitempty"`
			} `json:"parentCompany,omitempty"`
			TagOptions []struct {
				HelpText string `json:"helpText,omitempty"`
				ID       int64  `json:"id,omitempty"`
				Name     string `json:"name,omitempty"`
				Tag      int64  `json:"tag,omitempty"`
				TagName  string `json:"tagName,omitempty"`
			} `json:"tagOptions,omitempty"`
		} `json:"distributor,omitempty"`
		SpecialEffects []struct {
			ActiveDate string `json:"activeDate,omitempty"`
			Aliases    []struct {
				Language string `json:"language,omitempty"`
				Name     string `json:"name,omitempty"`
			} `json:"aliases,omitempty"`
			Country              string   `json:"country,omitempty"`
			ID                   int64    `json:"id,omitempty"`
			InactiveDate         string   `json:"inactiveDate,omitempty"`
			Name                 string   `json:"name,omitempty"`
			NameTranslations     []string `json:"nameTranslations,omitempty"`
			OverviewTranslations []string `json:"overviewTranslations,omitempty"`
			PrimaryCompanyType   int64    `json:"primaryCompanyType,omitempty"`
			Slug                 string   `json:"slug,omitempty"`
			ParentCompany        struct {
				ID       int64  `json:"id,omitempty"`
				Name     string `json:"name,omitempty"`
				Relation struct {
					ID       int64  `json:"id,omitempty"`
					TypeName string `json:"typeName,omitempty"`
				} `json:"relation,omitempty"`
			} `json:"parentCompany,omitempty"`
			TagOptions []struct {
				HelpText string `json:"helpText,omitempty"`
				ID       int64  `json:"id,omitempty"`
				Name     string `json:"name,omitempty"`
				Tag      int64  `json:"tag,omitempty"`
				TagName  string `json:"tagName,omitempty"`
			} `json:"tagOptions,omitempty"`
		} `json:"special_effects,omitempty"`
	} `json:"companies,omitempty"`
	SeriesId int64 `json:"seriesId,omitempty"`
	Type     struct {
		AlternateName string `json:"alternateName,omitempty"`
		ID            int64  `json:"id,omitempty"`
		Name          string `json:"name,omitempty"`
		Type          string `json:"type,omitempty"`
	} `json:"type,omitempty"`
	Year string `json:"year,omitempty"`
}

type ExtendedSeriesResponse struct { // INCOMPLETE
	Data *struct {
		Seasons []*ExtendedSeriesResponse_Season `json:"seasons,omitempty"`
	} `json:"data,omitempty"`
	Status string `json:"status,omitempty"`
}

// +--------------------------------------------------------------------------------------------+

//goland:noinspection GoSnakeCaseUsage
type ExtendedSeasonsResponse_Episode struct {
	Aired                string   `json:"aired"`
	AirsAfterSeason      int64    `json:"airsAfterSeason"`
	AirsBeforeEpisode    int64    `json:"airsBeforeEpisode"`
	AirsBeforeSeason     int64    `json:"airsBeforeSeason"`
	FinaleType           string   `json:"finaleType"`
	ID                   int64    `json:"id"`
	Image                string   `json:"image"`
	ImageType            int64    `json:"imageType"`
	IsMovie              int64    `json:"isMovie"`
	LastUpdated          string   `json:"lastUpdated"`
	LinkedMovie          int64    `json:"linkedMovie"`
	Name                 string   `json:"name"`
	NameTranslations     []string `json:"nameTranslations"`
	Number               int64    `json:"number"`
	Overview             string   `json:"overview"`
	OverviewTranslations []string `json:"overviewTranslations"`
	Runtime              int64    `json:"runtime"`
	SeasonNumber         int64    `json:"seasonNumber"`
	Seasons              []struct {
		ID                   int64    `json:"id"`
		Image                string   `json:"image"`
		ImageType            int64    `json:"imageType"`
		LastUpdated          string   `json:"lastUpdated"`
		Name                 string   `json:"name"`
		NameTranslations     []string `json:"nameTranslations"`
		Number               int64    `json:"number"`
		OverviewTranslations []string `json:"overviewTranslations"`
		Companies            struct {
			Studio []struct {
				ActiveDate string `json:"activeDate"`
				Aliases    []struct {
					Language string `json:"language"`
					Name     string `json:"name"`
				} `json:"aliases"`
				Country              string   `json:"country"`
				ID                   int64    `json:"id"`
				InactiveDate         string   `json:"inactiveDate"`
				Name                 string   `json:"name"`
				NameTranslations     []string `json:"nameTranslations"`
				OverviewTranslations []string `json:"overviewTranslations"`
				PrimaryCompanyType   int64    `json:"primaryCompanyType"`
				Slug                 string   `json:"slug"`
				ParentCompany        struct {
					ID       int64  `json:"id"`
					Name     string `json:"name"`
					Relation struct {
						ID       int64  `json:"id"`
						TypeName string `json:"typeName"`
					} `json:"relation"`
				} `json:"parentCompany"`
				TagOptions []struct {
					HelpText string `json:"helpText"`
					ID       int64  `json:"id"`
					Name     string `json:"name"`
					Tag      int64  `json:"tag"`
					TagName  string `json:"tagName"`
				} `json:"tagOptions"`
			} `json:"studio"`
			Network []struct {
				ActiveDate string `json:"activeDate"`
				Aliases    []struct {
					Language string `json:"language"`
					Name     string `json:"name"`
				} `json:"aliases"`
				Country              string   `json:"country"`
				ID                   int64    `json:"id"`
				InactiveDate         string   `json:"inactiveDate"`
				Name                 string   `json:"name"`
				NameTranslations     []string `json:"nameTranslations"`
				OverviewTranslations []string `json:"overviewTranslations"`
				PrimaryCompanyType   int64    `json:"primaryCompanyType"`
				Slug                 string   `json:"slug"`
				ParentCompany        struct {
					ID       int64  `json:"id"`
					Name     string `json:"name"`
					Relation struct {
						ID       int64  `json:"id"`
						TypeName string `json:"typeName"`
					} `json:"relation"`
				} `json:"parentCompany"`
				TagOptions []struct {
					HelpText string `json:"helpText"`
					ID       int64  `json:"id"`
					Name     string `json:"name"`
					Tag      int64  `json:"tag"`
					TagName  string `json:"tagName"`
				} `json:"tagOptions"`
			} `json:"network"`
			Production []struct {
				ActiveDate string `json:"activeDate"`
				Aliases    []struct {
					Language string `json:"language"`
					Name     string `json:"name"`
				} `json:"aliases"`
				Country              string   `json:"country"`
				ID                   int64    `json:"id"`
				InactiveDate         string   `json:"inactiveDate"`
				Name                 string   `json:"name"`
				NameTranslations     []string `json:"nameTranslations"`
				OverviewTranslations []string `json:"overviewTranslations"`
				PrimaryCompanyType   int64    `json:"primaryCompanyType"`
				Slug                 string   `json:"slug"`
				ParentCompany        struct {
					ID       int64  `json:"id"`
					Name     string `json:"name"`
					Relation struct {
						ID       int64  `json:"id"`
						TypeName string `json:"typeName"`
					} `json:"relation"`
				} `json:"parentCompany"`
				TagOptions []struct {
					HelpText string `json:"helpText"`
					ID       int64  `json:"id"`
					Name     string `json:"name"`
					Tag      int64  `json:"tag"`
					TagName  string `json:"tagName"`
				} `json:"tagOptions"`
			} `json:"production"`
			Distributor []struct {
				ActiveDate string `json:"activeDate"`
				Aliases    []struct {
					Language string `json:"language"`
					Name     string `json:"name"`
				} `json:"aliases"`
				Country              string   `json:"country"`
				ID                   int64    `json:"id"`
				InactiveDate         string   `json:"inactiveDate"`
				Name                 string   `json:"name"`
				NameTranslations     []string `json:"nameTranslations"`
				OverviewTranslations []string `json:"overviewTranslations"`
				PrimaryCompanyType   int64    `json:"primaryCompanyType"`
				Slug                 string   `json:"slug"`
				ParentCompany        struct {
					ID       int64  `json:"id"`
					Name     string `json:"name"`
					Relation struct {
						ID       int64  `json:"id"`
						TypeName string `json:"typeName"`
					} `json:"relation"`
				} `json:"parentCompany"`
				TagOptions []struct {
					HelpText string `json:"helpText"`
					ID       int64  `json:"id"`
					Name     string `json:"name"`
					Tag      int64  `json:"tag"`
					TagName  string `json:"tagName"`
				} `json:"tagOptions"`
			} `json:"distributor"`
			SpecialEffects []struct {
				ActiveDate string `json:"activeDate"`
				Aliases    []struct {
					Language string `json:"language"`
					Name     string `json:"name"`
				} `json:"aliases"`
				Country              string   `json:"country"`
				ID                   int64    `json:"id"`
				InactiveDate         string   `json:"inactiveDate"`
				Name                 string   `json:"name"`
				NameTranslations     []string `json:"nameTranslations"`
				OverviewTranslations []string `json:"overviewTranslations"`
				PrimaryCompanyType   int64    `json:"primaryCompanyType"`
				Slug                 string   `json:"slug"`
				ParentCompany        struct {
					ID       int64  `json:"id"`
					Name     string `json:"name"`
					Relation struct {
						ID       int64  `json:"id"`
						TypeName string `json:"typeName"`
					} `json:"relation"`
				} `json:"parentCompany"`
				TagOptions []struct {
					HelpText string `json:"helpText"`
					ID       int64  `json:"id"`
					Name     string `json:"name"`
					Tag      int64  `json:"tag"`
					TagName  string `json:"tagName"`
				} `json:"tagOptions"`
			} `json:"special_effects"`
		} `json:"companies"`
		SeriesID int64 `json:"seriesId"`
		Type     struct {
			AlternateName string `json:"alternateName"`
			ID            int64  `json:"id"`
			Name          string `json:"name"`
			Type          string `json:"type"`
		} `json:"type"`
		Year string `json:"year"`
	} `json:"seasons"`
	SeriesID   int64  `json:"seriesId"`
	SeasonName string `json:"seasonName"`
	Year       string `json:"year"`
}

type ExtendedSeasonsResponse struct {
	Data *struct {
		Episodes []*ExtendedSeasonsResponse_Episode `json:"episodes,omitempty"`
	} `json:"data,omitempty"`
	Status string `json:"status,omitempty"`
}
