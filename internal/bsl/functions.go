package bsl

import "strings"

// Function describes a built-in BSL function.
type Function struct {
	Name        string
	NameEn      string
	Description string
	Syntax      string
	Parameters  string
	ReturnType  string
	Example     string
}

// BuiltinFunctions is the reference of 1C platform built-in functions.
var BuiltinFunctions = []Function{
	{
		Name:        "СтрНайти",
		NameEn:      "StrFind",
		Description: "Находит вхождение искомой строки в строке-источнике.",
		Syntax:      "СтрНайти(<Строка>, <ПодстрокаПоиска>, <НаправлениеПоиска>, <НачальнаяПозиция>, <НомерВхождения>)",
		Parameters:  "Строка (обязат.), ПодстрокаПоиска (обязат.), НаправлениеПоиска (необязат.), НачальнаяПозиция (необязат.), НомерВхождения (необязат.)",
		ReturnType:  "Число",
		Example:     `Позиция = СтрНайти("Hello World", "World"); // 7`,
	},
	{
		Name:        "Формат",
		NameEn:      "Format",
		Description: "Формирует удобочитаемое представление значения по строке форматирования.",
		Syntax:      "Формат(<Значение>, <ФорматнаяСтрока>)",
		Parameters:  "Значение (обязат.), ФорматнаяСтрока (обязат.)",
		ReturnType:  "Строка",
		Example:     `Формат(1234.56, "ЧДЦ=2"); // "1 234,56"`,
	},
	{
		Name:        "Запрос",
		NameEn:      "Query",
		Description: "Объект для выполнения запросов к базе данных на языке запросов 1С.",
		Syntax:      "Запрос = Новый Запрос(<ТекстЗапроса>)",
		Parameters:  "ТекстЗапроса (необязат.)",
		ReturnType:  "Запрос",
		Example:     "Запрос = Новый Запрос(\"ВЫБРАТЬ Наименование ИЗ Справочник.Контрагенты\");\nРезультат = Запрос.Выполнить();",
	},
}

// Search finds functions by name (Russian or English), case-insensitive.
func Search(query string) []Function {
	queryLower := strings.ToLower(query)
	var results []Function
	for _, f := range BuiltinFunctions {
		if strings.Contains(strings.ToLower(f.Name), queryLower) || strings.Contains(strings.ToLower(f.NameEn), queryLower) {
			results = append(results, f)
		}
	}
	return results
}
