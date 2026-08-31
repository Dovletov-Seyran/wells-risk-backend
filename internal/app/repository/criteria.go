package repository

import "time"

func createdOn(day int) time.Time {
	return time.Date(2026, time.August, day, 9, 0, 0, 0, time.UTC)
}

func seedCriteria() []WellsCriterion {
	return []WellsCriterion{
		{
			CriterionID: 1, CriterionName: "Активное онкологическое заболевание",
			ShortDescription: "Лечение или паллиативная помощь в течение последних 6 месяцев.",
			WellsPoints:      1, ScaleType: "ТГВ", CriterionGroup: "анамнез",
			AssessmentMethod: "опрос пациента и медицинская документация",
			ImageKey:         "active_cancer", VideoKey: "active_cancer",
			CriterionStatus: StatusPublished, CreatedAt: createdOn(10), LikedByPhysicians: []int{1, 2, 4},
		},
		{
			CriterionID: 2, CriterionName: "Постельный режим или операция за 12 недель",
			ShortDescription: "Учитывается операция под общей или регионарной анестезией.",
			WellsPoints:      1, ScaleType: "ТГВ", CriterionGroup: "анамнез",
			AssessmentMethod: "опрос пациента и медицинская документация",
			ImageKey:         "bedridden_surgery", VideoKey: "bedridden_surgery",
			CriterionStatus: StatusPublished, CreatedAt: createdOn(10), LikedByPhysicians: []int{2},
		},
		{
			CriterionID: 3, CriterionName: "Увеличение окружности голени более 3 см",
			ShortDescription: "Замер на 10 см ниже бугристости большеберцовой кости, сравнение со здоровой ногой.",
			WellsPoints:      1, ScaleType: "ТГВ", CriterionGroup: "измерение",
			AssessmentMethod: "сантиметровая лента",
			ImageKey:         "calf_swelling", VideoKey: "calf_swelling",
			CriterionStatus: StatusPublished, CreatedAt: createdOn(11), LikedByPhysicians: []int{1, 3, 5, 6},
		},
		{
			CriterionID: 4, CriterionName: "Расширенные коллатеральные поверхностные вены",
			ShortDescription: "Неварикозные вены, появившиеся на поражённой конечности.",
			WellsPoints:      1, ScaleType: "ТГВ", CriterionGroup: "осмотр",
			AssessmentMethod: "визуальный осмотр конечности",
			ImageKey:         "collateral_veins", VideoKey: "collateral_veins",
			CriterionStatus: StatusPublished, CreatedAt: createdOn(11), LikedByPhysicians: []int{4},
		},
		{
			CriterionID: 5, CriterionName: "Отёк всей нижней конечности",
			ShortDescription: "Отёк захватывает конечность целиком, а не только голень.",
			WellsPoints:      1, ScaleType: "ТГВ", CriterionGroup: "осмотр",
			AssessmentMethod: "визуальный осмотр конечности",
			ImageKey:         "entire_leg_swollen", VideoKey: "entire_leg_swollen",
			CriterionStatus: StatusPublished, CreatedAt: createdOn(12), LikedByPhysicians: []int{1, 5},
		},
		{
			CriterionID: 6, CriterionName: "Локальная болезненность по ходу глубоких вен",
			ShortDescription: "Болезненность при пальпации по проекции глубокой венозной системы.",
			WellsPoints:      1, ScaleType: "ТГВ", CriterionGroup: "пальпация",
			AssessmentMethod: "пальпация конечности",
			ImageKey:         "localized_tenderness", VideoKey: "localized_tenderness",
			CriterionStatus: StatusPublished, CreatedAt: createdOn(12), LikedByPhysicians: []int{2, 3},
		},
		{
			CriterionID: 7, CriterionName: "Отёк с ямкой при надавливании на симптомной ноге",
			ShortDescription: "Ямка сохраняется после надавливания только на поражённой конечности.",
			WellsPoints:      1, ScaleType: "ТГВ", CriterionGroup: "пальпация",
			AssessmentMethod: "пальпация конечности",
			ImageKey:         "pitting_edema", VideoKey: "pitting_edema",
			CriterionStatus: StatusPublished, CreatedAt: createdOn(13), LikedByPhysicians: []int{6},
		},
		{
			CriterionID: 8, CriterionName: "Паралич, парез или недавняя иммобилизация конечности",
			ShortDescription: "Учитывается в том числе гипсовая иммобилизация нижней конечности.",
			WellsPoints:      1, ScaleType: "ТГВ", CriterionGroup: "осмотр",
			AssessmentMethod: "неврологический осмотр",
			ImageKey:         "paralysis_immobilization", VideoKey: "paralysis_immobilization",
			CriterionStatus: StatusPublished, CreatedAt: createdOn(13), LikedByPhysicians: []int{1, 2, 3},
		},
		{
			CriterionID: 9, CriterionName: "Ранее документированный ТГВ",
			ShortDescription: "Подтверждённый инструментально эпизод тромбоза глубоких вен в анамнезе.",
			WellsPoints:      1, ScaleType: "ТГВ", CriterionGroup: "анамнез",
			AssessmentMethod: "медицинская документация",
			ImageKey:         "previous_dvt", VideoKey: "previous_dvt",
			CriterionStatus: StatusPublished, CreatedAt: createdOn(14), LikedByPhysicians: []int{5},
		},
		{
			CriterionID: 10, CriterionName: "Альтернативный диагноз не менее вероятен, чем ТГВ",
			ShortDescription: "Единственный критерий с отрицательным весом: снижает итоговую сумму баллов.",
			WellsPoints:      -2, ScaleType: "ТГВ", CriterionGroup: "клиническая оценка",
			AssessmentMethod: "дифференциальная диагностика",
			ImageKey:         "alternative_diagnosis", VideoKey: "alternative_diagnosis",
			CriterionStatus: StatusPublished, CreatedAt: createdOn(14), LikedByPhysicians: []int{1, 2, 3, 4, 5},
		},
		{
			CriterionID: 11, CriterionName: "Клинические признаки и симптомы ТГВ",
			ShortDescription: "Отёк конечности и болезненность при пальпации глубоких вен.",
			WellsPoints:      3, ScaleType: "ТЭЛА", CriterionGroup: "осмотр",
			AssessmentMethod: "осмотр нижних конечностей",
			ImageKey:         "dvt_signs", VideoKey: "dvt_signs",
			CriterionStatus: StatusPublished, CreatedAt: createdOn(15), LikedByPhysicians: []int{1, 2, 3, 4},
		},
		{
			CriterionID: 12, CriterionName: "ТЭЛА - наиболее вероятный диагноз",
			ShortDescription: "ТЭЛА вероятнее альтернатив или равновероятна им.",
			WellsPoints:      3, ScaleType: "ТЭЛА", CriterionGroup: "клиническая оценка",
			AssessmentMethod: "дифференциальная диагностика",
			ImageKey:         "pe_most_likely", VideoKey: "pe_most_likely",
			CriterionStatus: StatusPublished, CreatedAt: createdOn(15), LikedByPhysicians: []int{2, 6},
		},
		{
			CriterionID: 13, CriterionName: "ЧСС более 100 ударов в минуту",
			ShortDescription: "Синусовая тахикардия в покое на момент осмотра.",
			WellsPoints:      1.5, ScaleType: "ТЭЛА", CriterionGroup: "инструментальное измерение",
			AssessmentMethod: "пульсоксиметрия, ЭКГ",
			ImageKey:         "heart_rate_over_100", VideoKey: "heart_rate_over_100",
			CriterionStatus: StatusPublished, CreatedAt: createdOn(16), LikedByPhysicians: []int{3, 4, 5},
		},
		{
			CriterionID: 14, CriterionName: "Иммобилизация от 3 дней или операция за 4 недели",
			ShortDescription: "Окно для шкалы ТЭЛА короче, чем для шкалы ТГВ: 4 недели вместо 12.",
			WellsPoints:      1.5, ScaleType: "ТЭЛА", CriterionGroup: "анамнез",
			AssessmentMethod: "опрос пациента и медицинская документация",
			ImageKey:         "immobilization_surgery", VideoKey: "immobilization_surgery",
			CriterionStatus: StatusPublished, CreatedAt: createdOn(16), LikedByPhysicians: []int{1},
		},
		{
			CriterionID: 15, CriterionName: "Ранее подтверждённые ТЭЛА или ТГВ",
			ShortDescription: "Объективно подтверждённый эпизод венозной тромбоэмболии в анамнезе.",
			WellsPoints:      1.5, ScaleType: "ТЭЛА", CriterionGroup: "анамнез",
			AssessmentMethod: "медицинская документация",
			ImageKey:         "previous_vte", VideoKey: "previous_vte",
			CriterionStatus: StatusPublished, CreatedAt: createdOn(17), LikedByPhysicians: []int{2, 5},
		},
		{
			CriterionID: 16, CriterionName: "Кровохарканье",
			ShortDescription: "Примесь крови в мокроте на фоне одышки или боли в груди.",
			WellsPoints:      1, ScaleType: "ТЭЛА", CriterionGroup: "жалоба",
			AssessmentMethod: "опрос пациента",
			ImageKey:         "hemoptysis", VideoKey: "hemoptysis",
			CriterionStatus: StatusPublished, CreatedAt: createdOn(17), LikedByPhysicians: []int{4, 6},
		},
		{
			CriterionID: 17, CriterionName: "Злокачественное новообразование в течение 6 месяцев",
			ShortDescription: "", WellsPoints: 0, ScaleType: "", CriterionGroup: "", AssessmentMethod: "",
			ImageKey: "malignancy", VideoKey: "malignancy",
			CriterionStatus: StatusDraft, CreatedAt: createdOn(18), LikedByPhysicians: []int{},
		},
		{
			CriterionID: 18, CriterionName: "Признаки ТГВ (дубль записи)",
			ShortDescription: "Ошибочно созданная вторая запись того же критерия.",
			WellsPoints:      3, ScaleType: "ТЭЛА", CriterionGroup: "осмотр",
			AssessmentMethod: "осмотр нижних конечностей",
			ImageKey:         "dvt_signs", VideoKey: "dvt_signs",
			CriterionStatus: StatusDeleted, CreatedAt: createdOn(18), LikedByPhysicians: []int{},
		},
	}
}
