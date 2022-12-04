package backend

import "salespalm/server/app/entities"

type ITemplateCompilerService interface {
	Commons() *TemplateCompilerCommons
}

type TemplateCompilerServiceImpl struct {
	ITemplateCompilerService

	commons TemplateCompilerCommons
}

type TemplateCompilerCommons struct {
	Variables   []*Variable
	StubContact *entities.Contact
}

func (c *TemplateCompilerServiceImpl) Commons() *TemplateCompilerCommons {
	return &c.commons
}

func (c *TemplateCompilerServiceImpl) Init() {
	c.commons.Variables = c.variables()
	c.commons.StubContact = c.stubContact()
}

func (c *TemplateCompilerServiceImpl) variables() []*Variable {
	return []*Variable{
		{
			Name:        "Contact.FirstName",
			Description: "Имя",
		},
		{
			Name:        "Contact.LastName",
			Description: "Фамилия",
		},
		{
			Name:        "Contact.Linkedin",
			Description: "Linkedin",
		},
		{
			Name:        "Contact.Phone",
			Description: "Телефон",
		},
		{
			Name:        "Contact.Email",
			Description: "Email",
		},
		{
			Name:        "Contact.Company",
			Description: "Компания",
		},
		{
			Name:        "Me.FirstName",
			Description: "Мое имя",
		},
		{
			Name:        "Me.LastName",
			Description: "Моя фамилия",
		},
		{
			Name:        "Me.Linkedin",
			Description: "Мой Linkedin",
		},
		{
			Name:        "Me.Phone",
			Description: "Мой телефон",
		},
		{
			Name:        "Me.Email",
			Description: "Мой Email",
		},
		{
			Name:        "Me.Company",
			Description: "Моя компания",
		},

		{
			Name:        "Sequence.Sendings",
			Description: "Сколько раз контакт получил мое сообщение",
		},
		{
			Name:        "Sequence.Views",
			Description: "Сколько раз контакт посмотрел мое сообщение",
		},
	}
}

func (c *TemplateCompilerServiceImpl) stubContact() *entities.Contact {
	return &entities.Contact{
		Phone:     "+7 (999) 11-22-33",
		FirstName: "Александр",
		LastName:  "Васнецов",
		Email:     "alex1987@gmail.com",
		Company:   `ООО "Инвестброкер"`,
		Linkedin:  "https://www.linkedin.ru/alex1987",
	}
}

type Variable struct {
	Name, Description string
}
