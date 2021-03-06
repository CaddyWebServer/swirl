package compose

import "github.com/pkg/errors"

// Interpolate replaces variables in a string with the values from a mapping
func Interpolate(config map[string]interface{}, section string, mapping Mapping) (map[string]interface{}, error) {
	out := map[string]interface{}{}

	for name, item := range config {
		if item == nil {
			out[name] = nil
			continue
		}
		mapItem, ok := item.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("Invalid type for %s : %T instead of %T", name, item, out)
		}
		interpolatedItem, err := interpolateSectionItem(name, mapItem, section, mapping)
		if err != nil {
			return nil, err
		}
		out[name] = interpolatedItem
	}

	return out, nil
}

func interpolateSectionItem(
	name string,
	item map[string]interface{},
	section string,
	mapping Mapping,
) (map[string]interface{}, error) {

	out := map[string]interface{}{}

	for key, value := range item {
		interpolatedValue, err := recursiveInterpolate(value, mapping)
		switch err := err.(type) {
		case nil:
		case *InvalidTemplateError:
			return nil, errors.Errorf(
				"Invalid interpolation format for %#v option in %s %#v: %#v. You may need to escape any $ with another $.",
				key, section, name, err.Template,
			)
		default:
			return nil, errors.Wrapf(err, "error while interpolating %s in %s %s", key, section, name)
		}
		out[key] = interpolatedValue
	}

	return out, nil

}

func recursiveInterpolate(
	value interface{},
	mapping Mapping,
) (interface{}, error) {

	switch value := value.(type) {

	case string:
		return Substitute(value, mapping)

	case map[string]interface{}:
		out := map[string]interface{}{}
		for key, elem := range value {
			interpolatedElem, err := recursiveInterpolate(elem, mapping)
			if err != nil {
				return nil, err
			}
			out[key] = interpolatedElem
		}
		return out, nil

	case []interface{}:
		out := make([]interface{}, len(value))
		for i, elem := range value {
			interpolatedElem, err := recursiveInterpolate(elem, mapping)
			if err != nil {
				return nil, err
			}
			out[i] = interpolatedElem
		}
		return out, nil

	default:
		return value, nil

	}

}
