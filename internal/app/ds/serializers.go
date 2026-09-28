package ds

import (
	"strings"
	"time"
)

var filesBaseURL = "http://localhost:9000/wells-criteria"

func SetFilesBaseURL(baseURL string) {
	if baseURL != "" {
		filesBaseURL = strings.TrimRight(baseURL, "/")
	}
}

func FileURL(objectKey string) string {
	if objectKey == "" {
		return ""
	}

	return filesBaseURL + "/" + objectKey
}

type CriterionResponse struct {
	CriterionID      int        `json:"criterion_id"`
	CriterionName    string     `json:"criterion_name"`
	ShortDescription string     `json:"short_description"`
	CriterionGroup   string     `json:"criterion_group"`
	WellsPoints      float64    `json:"wells_points"`
	ImageURL         string     `json:"image_url"`
	VideoURL         string     `json:"video_url"`
	LikeCount        int        `json:"like_count"`
	IsLiked          int        `json:"is_liked"`
	IsMine           int        `json:"is_mine"`
	CreatedAt        time.Time  `json:"created_at"`
	FormedAt         *time.Time `json:"formed_at"`
}

func NewCriterionResponse(criterion WellsCriterion, currentPhysicianID int) CriterionResponse {
	isMine := 0
	if criterion.CreatorID != nil && *criterion.CreatorID == currentPhysicianID {
		isMine = 1
	}

	isLiked := 0
	if criterion.LikedByMe {
		isLiked = 1
	}

	return CriterionResponse{
		CriterionID:      criterion.CriterionID,
		CriterionName:    criterion.CriterionName,
		ShortDescription: criterion.ShortDescription,
		CriterionGroup:   criterion.CriterionGroup,
		WellsPoints:      criterion.WellsPoints,
		ImageURL:         FileURL(criterion.ImageKey),
		VideoURL:         FileURL(criterion.VideoKey),
		LikeCount:        criterion.LikeCount,
		IsLiked:          isLiked,
		IsMine:           isMine,
		CreatedAt:        criterion.CreatedAt,
		FormedAt:         criterion.FormedAt,
	}
}

func NewCriterionListResponse(criteria []WellsCriterion, currentPhysicianID int) []CriterionResponse {
	list := make([]CriterionResponse, 0, len(criteria))

	for _, criterion := range criteria {
		list = append(list, NewCriterionResponse(criterion, currentPhysicianID))
	}

	return list
}

type PhysicianResponse struct {
	PhysicianID int    `json:"physician_id"`
	Login       string `json:"login"`
	FullName    string `json:"full_name"`
	IsModerator bool   `json:"is_moderator"`
}

// NewPhysicianResponse сериализует врача.
func NewPhysicianResponse(physician Physician) PhysicianResponse {
	return PhysicianResponse{
		PhysicianID: physician.PhysicianID,
		Login:       physician.Login,
		FullName:    physician.FullName,
		IsModerator: physician.IsModerator,
	}
}
