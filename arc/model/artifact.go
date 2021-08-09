package model

import "fmt"

type Artifacts struct {
	// Values contains all the value objects.
	Values []NamedWithFields `tadl:"value"`
	// Entities contains all defined entities.
	// Entities differ from value objects as they are different from each
	// other even if they have the same values, like people with the
	// same name. The will get a unique Id when generated.
	Entities   []NamedWithFields `tadl:"entity"`
	Aggregates []Aggregate       `tadl:"aggregate"`
}

func (a *Artifacts) Validate(project *Project) error {
	for _, aggregate := range a.Aggregates {
		if err := aggregate.Validate(project); err != nil {
			return err
		}
	}

	return nil
}

// Aggregate is a combination of several value objects and entities.
type Aggregate struct {
	// Name is the name of this aggregate.
	Name string `tadl:"name,attr"`
	// Fields describes what type should be stored in this aggregate.
	Fields map[string]string `tadl:"fields"`
	// Methods are operations that are defined for this aggregate.
	// The methods name is the map key.
	Methods map[string]Method `tadl:"methods"`
}

func (a *Aggregate) Validate(project *Project) error {
	for methodName, method := range a.Methods {
		foundAny := false

		for _, story := range project.Stories {
			if method.Story == story.Id {
				foundAny = true
				break
			}
		}

		if !foundAny {
			// TODO Proper error types
			return fmt.Errorf("aggregate '%s' method '%s' references story with ID %d, but such a story does not exist", a.Name, methodName, method.Story)
		}
	}

	return nil
}

// Method has a description and input and output parameters.
// Both parameters are maps from name to type.
type Method struct {
	Description string `tadl:",inner"`
	// Because is a reference to the ID of a user story that explains why
	// this method should exist.
	Story   int               `tadl:"story,attr"`
	Params  map[string]string `tadl:",inner"`
	Returns map[string]string `tadl:"ret"`
}

// NamedWithFields can be a value object or an entity.
// It has a name, a description and several "name: type" fields.
type NamedWithFields struct {
	Name        string            `tadl:"name,attr"`
	Description string            `tadl:",inner"`
	Fields      map[string]string `tadl:",inner"`
}
