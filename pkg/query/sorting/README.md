<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# sorting

```go
import "github.com/greenbone/opensight-golang-libraries/pkg/query/sorting"
```

## Index

- [Constants](<#constants>)
- [func AddRequest\(transaction \*gorm.DB, params Params\) \*gorm.DB](<#AddRequest>)
- [func NewSortingError\(format string, value ...any\) error](<#NewSortingError>)
- [func ValidateSortingRequest\(req \*Request\) error](<#ValidateSortingRequest>)
- [type Error](<#Error>)
  - [func \(e \*Error\) Error\(\) string](<#Error.Error>)
- [type Params](<#Params>)
  - [func DetermineEffectiveSortingParams\(model SortingSettingsInterface, sortingReq \*Request\) \(Params, error\)](<#DetermineEffectiveSortingParams>)
- [type Request](<#Request>)
- [type Response](<#Response>)
- [type SortDefault](<#SortDefault>)
  - [func GetSortDefaults\(model SortingSettingsInterface\) \(result SortDefault, err error\)](<#GetSortDefaults>)
- [type SortDirection](<#SortDirection>)
  - [func SortDirectionFromString\(str string\) SortDirection](<#SortDirectionFromString>)
  - [func \(s SortDirection\) String\(\) string](<#SortDirection.String>)
- [type SortableColumn](<#SortableColumn>)
  - [func GetSortableColumns\(model SortingSettingsInterface\) \(sortables \[\]SortableColumn\)](<#GetSortableColumns>)
- [type SortingSettingsInterface](<#SortingSettingsInterface>)


## Constants

<a name="SortColumnTag"></a>

```go
const (
    // SortColumnTag is a tag to define the field name to be used for sorting
    SortColumnTag = "sortColumn"
    /*SortColumnOverrideTag is a tag to override the field name to be used for sorting.

    An example use case is the following:

    If you take a column to order by, which is part of the original table (like hostname) the SQL syntax is fine - see the generated GORM query below.´.
    SELECT "asset"."id","asset"."created_at","asset"."updated_at","asset"."greenbone_agent_id","asset"."last_authenticated_scan_at","asset"."mac_address","asset"."net_bios_name","asset"."ssh_fingerprint","asset"."has_agent","asset"."has_vt_result","asset"."hostname","asset"."ip","asset"."last_scan_at","asset"."operating_system","asset"."deleted_at","asset"."deleted_by","asset"."source_id","asset"."appliance_id" FROM "asset" LEFT JOIN "appliance" "Appliance" ON "asset"."appliance_id" = "Appliance"."id" LEFT JOIN "installed_software" "InstalledSoftwares" ON "asset"."id" = "InstalledSoftwares"."asset_id" WHERE ("Appliance"."name" ILIKE '%Example%') ORDER BY hostname ASC LIMIT 10

    If we now change to sort by a column of a joined table, the SQL syntax is "TABLENAME"."Columnname". Now see the
    LEFT JOIN "appliance" "Appliance" which means to access the table appliance we now need to use "Appliance" instead of "appliance".
    Therefore its necassary to change the SQL syntax for the foreign table sort.
    SELECT "asset"."id","asset"."created_at","asset"."updated_at","asset"."greenbone_agent_id","asset"."last_authenticated_scan_at","asset"."mac_address","asset"."net_bios_name","asset"."ssh_fingerprint","asset"."has_agent","asset"."has_vt_result","asset"."hostname","asset"."ip","asset"."last_scan_at","asset"."operating_system","asset"."deleted_at","asset"."deleted_by","asset"."source_id","asset"."appliance_id" FROM "asset" LEFT JOIN "appliance" "Appliance" ON "asset"."appliance_id" = "Appliance"."id" LEFT JOIN "installed_software" "InstalledSoftwares" ON "asset"."id" = "InstalledSoftwares"."asset_id" ORDER BY "Appliance"."name" DESC LIMIT 20

    If we query for the appliance name GORM changes it correctly to "Appliance"."name" - see the query below. And the same we now do in our code for the sorting - we change the JOINED field from appliance.name to "Appliance"."name".
    SELECT "asset"."id","asset"."created_at","asset"."updated_at","asset"."greenbone_agent_id","asset"."last_authenticated_scan_at","asset"."mac_address","asset"."net_bios_name","asset"."ssh_fingerprint","asset"."has_agent","asset"."has_vt_result","asset"."hostname","asset"."ip","asset"."last_scan_at","asset"."operating_system","asset"."deleted_at","asset"."deleted_by","asset"."source_id","asset"."appliance_id" FROM "asset" LEFT JOIN "appliance" "Appliance" ON "asset"."appliance_id" = "Appliance"."id" LEFT JOIN "installed_software" "InstalledSoftwares" ON "asset"."id" = "InstalledSoftwares"."asset_id" WHERE ("Appliance"."name" ILIKE '%Example%') ORDER BY hostname ASC LIMIT 10
    */
    SortColumnOverrideTag = "sortColumnOverride"
    // SortDirectionTag is a tag that defines the sort (i.e. direction of sorting); must be a SortDirection
    SortDirectionTag = "sortDirection"
)
```

<a name="AddRequest"></a>
## func AddRequest

```go
func AddRequest(transaction *gorm.DB, params Params) *gorm.DB
```



<a name="NewSortingError"></a>
## func NewSortingError

```go
func NewSortingError(format string, value ...any) error
```



<a name="ValidateSortingRequest"></a>
## func ValidateSortingRequest

```go
func ValidateSortingRequest(req *Request) error
```

ValidateSortingRequest validates a sorting request.

<a name="Error"></a>
## type Error



```go
type Error struct {
    Msg string
}
```

<a name="Error.Error"></a>
### func \(\*Error\) Error

```go
func (e *Error) Error() string
```



<a name="Params"></a>
## type Params



```go
type Params struct {
    OriginalSortColumn  string
    SortDirection       SortDirection
    EffectiveSortColumn string
}
```

<a name="DetermineEffectiveSortingParams"></a>
### func DetermineEffectiveSortingParams

```go
func DetermineEffectiveSortingParams(model SortingSettingsInterface, sortingReq *Request) (Params, error)
```

DetermineEffectiveSortingParams checks the requested sorting and sets the defaults in case of an error. If a SortColumnOverrideTag \(sortColumnOverride\) is given, it's value will be used for sorting instead of SortColumnTag \(sortColumn\). For a detailed explanation see SortColumnOverrideTag

<a name="Request"></a>
## type Request

Request represents a sorting request with a specified sort column and sort direction.

Fields: \- SortColumn: the column to sort on \- SortDirection: the direction of sorting \(asc or desc\)

```go
type Request struct {
    SortColumn    string        `json:"column"`
    SortDirection SortDirection `json:"direction"`
}
```

<a name="Response"></a>
## type Response

Response represents the response structure for sorting column and direction. SortingColumn stores the name of the column which was used for sorting. SortingDirection stores the direction which was applied by the sorting.

```go
type Response struct {
    SortingColumn    string        `json:"column"`
    SortingDirection SortDirection `json:"direction"`
}
```

<a name="SortDefault"></a>
## type SortDefault

SortDefault holds the default for sort direction and sorting field.

```go
type SortDefault struct {
    Column    string
    Direction SortDirection
}
```

<a name="GetSortDefaults"></a>
### func GetSortDefaults

```go
func GetSortDefaults(model SortingSettingsInterface) (result SortDefault, err error)
```

GetSortDefaults returns the sortable fields based on the struct provided.

<a name="SortDirection"></a>
## type SortDirection



```go
type SortDirection string
```

<a name="DirectionDescending"></a>

```go
const (
    DirectionDescending SortDirection = "desc"
    DirectionAscending  SortDirection = "asc"
    NoDirection         SortDirection = ""
)
```

<a name="SortDirectionFromString"></a>
### func SortDirectionFromString

```go
func SortDirectionFromString(str string) SortDirection
```



<a name="SortDirection.String"></a>
### func \(SortDirection\) String

```go
func (s SortDirection) String() string
```



<a name="SortableColumn"></a>
## type SortableColumn

SortableColumn is a struct to hold the fields which the paging can sort by.

```go
type SortableColumn struct {
    Column         string
    ColumnOverride string
}
```

<a name="GetSortableColumns"></a>
### func GetSortableColumns

```go
func GetSortableColumns(model SortingSettingsInterface) (sortables []SortableColumn)
```

GetSortableColumns returns a list of sortable fields

<a name="SortingSettingsInterface"></a>
## type SortingSettingsInterface



```go
type SortingSettingsInterface interface {
    GetSortDefault() SortDefault
    GetSortingMap() map[string]string
    GetOverrideSortColumn(string) string
}
```

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)