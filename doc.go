/*
Package maptrans provides a generic map to manipulate dictionaries.

There are many cases where we have some JSON object and we want to convert it to
another JSON object. Since Go is strongly typed, a useful approach is to first
translate JSON object into a map from string to an interface. Such map is called
Maptrans. We can then define a translation of one Maptrans into another Maptrans.
Translations are defined by the specially constructed Maptrans.

Translation Types

Translating field to another field with a different name

The simplest case is when we take a field from one object and present it in the
result under a different name. In this case we just write the source and
destination strings.

Example of name conversions

	var someMap map[string]interface{} = map[string]interface{}{
		"uuid":               "UUID",
		"name":               "Name",
		"age":                "Age",
	}

Translating field to another field and changing value.

We can provide a function which will translate field value to another value.
The function can also perform some verification of the input. For this we need
to describe convewrsion using MapElement object which is defined as

    type MapElement struct {
	    TargetName     string               // Name of destination field
	    MapFunc        MapFunc              // Function that value to new value
	    ModFunc        ModFunc              // Function for object modification
	    Type           TranslationType      // Type of translation
	    Mandatory      bool                 // The field must be present if true
	    SubTranslation ObjectMapDescription // Subtranslation map for children
    }

There are several predefined MapFunc translators:

- IDMap translates any value to itself. It can be used to map common objects
to themselves.

- StringMap translates string to a string, trimming leading and trailing spaces.

- StringToLowerMap translates a string to lower-case string (and trims spaces).

- IdentifierMap does a string translation but rejects invalid identifiers.
An identifier should start with a letter or underscore and have only letters,
digits and underscores in it.

- IPAddrMap does a string translation of IP addresses which should be valid.

- CIDRMap does a string translation of IP addresses in a slash notation, e.g
. 1.2.3.4/24

- BoolMap converts boolean or string to a boolean.

- UUIDMap converts string to a string verifying that the source string is a
valid UUID

- StringArrayMap converts array of strings into another array of strings.

When Mandatory field is specified, the field must be present in the source
object.

Translating maps to maps

To translate one map into asnother, the Type should be specified as
ObjectTranslationn.  The SubTranslation is the translation specification for
the internal object.

Translating array of objects.

To translate an arary of objects into another array of objects, the Type
should be specified as ObjectArrayTranslation. The SubTranslation defines
translation for each element of an array.

Using values to modify the original objects.

Example JSON object

	{
	  "name": "myname"
	  "value": {
		 "fruit": "apple"
	  }
	}

If we want to present this as a "flat"  object

	{
	  "name": "myname"
	  "fruit": "apple"
	}

we need a ObjectArrayTranslation method.

Example

	var translationDescr = map[string]interface{}{
			"name": "Name"
		"uuid": maptrans.MapElement{
			TargetName: "UUID",
			Mandatory:  true,
			MapFunc:    maptrans.UUIDMap,
		},
		"alias": maptrans.MapElement{
			TargetName: "Alias",
			MapFunc:    maptrans.IdentifierMap,
		},
		"force": maptrans.MapElement{
			TargetName: "Force",
			MapFunc:    maptrans.BoolMap,
		},
		"info": maptrans.MapElement{
			TargetName: "Info",
			Mandatory:  true,
			Type:       maptrans.ObjectTranslation,
			SubTranslation: map[string]interface{}{
				"Port": maptrans.MapElement{
					Name:       "port",
					Mandatory:  false,
					MapFunc:    maptrans.IntegerMap,
				},
				"IPAddress": maptrans.MapElement{
					TargetName: "address",
					Mandatory:  true,
					MapFunc:    maptrans.CIDRMap,
				},
				"Route": maptrans.MapElement{
					TargetName: "route",
					Type:       maptrans.ObjectArrayTranslation,
					Mandatory:  false,
					SubTranslation: map[string]interface{}{
						"Destination": maptrans.MapElement{
							TargetName: "destination",
							Mandatory:  true,
							MapFunc:    maptrans.CIDRMap,
						},
						"Gateway": maptrans.MapElement{
							TargetName: "gateway",
							Mandatory:  true,
							MapFunc:    maptrans.IPAddrMap,
						},
					},
				},
			},
		},
	}


*/
package maptrans
