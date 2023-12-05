# Go Utility Functions

This repository contains utility functions for Go projects, enhancing the functionality of slice operations. These functions allow you to check for the presence of elements in a slice based on either a specific value or a filtering function.

## Functions

### 1. ContainsLambda

The `ContainsLambda` function is a generic function that checks if any element in a slice satisfies a specified filter function.

#### Signature

```go
func ContainsLambda[T any](elements []T, filterFunction func(element T) bool) bool
```

Parameters
elements []T: The slice to search through.
filterFunction func(element T) bool: The filter function used to test each element in the slice.
Return Value
Returns true if any element in the slice satisfies the filter function, otherwise returns false.
### 2. Contains
The `Contains` function checks if a slice contains a specific value. It works with any comparable type.


### Signature
```go
func Contains[T comparable](elements []T, value T) bool
```
Parameters
elements []T: The slice to search through.
value T: The value to search for in the slice.
Return Value
Returns true if the value is present in the slice, otherwise returns false.

### Usage Examples

#### Using ContainsLambda
```go
numbers := []int{1, 2, 3, 4, 5}
isEven := func(n int) bool { return n%2 == 0 }
fmt.Println(ContainsLambda(numbers, isEven)) // Output: true
```

#### Using Contains
```go
names := []string{"Alice", "Bob", "Charlie"}
fmt.Println(Contains(names, "Bob")) // Output: true
fmt.Println(Contains(names, "Daisy")) // Output: false
```

##### Notes
ContainsLambda is useful when you need to check for an element based on a complex condition or a custom comparison logic.
Contains provides a simpler and more direct way to check for the presence of a specific value.
These functions are generic and can work with slices of any type, but Contains requires the type to be comparable.


# MustParseTime(value string) time.Time

This function takes in a string value and returns a time.Time struct. It uses the time.Parse function to parse the input string value into a time.Time struct. If the parsing fails, it will panic with the error. This function is intended to be used in situations where the input string value is known to be a valid RFC3339 formatted date time string, and panicking is preferable to returning an error.

## Example Usage

```go
t, err := time.Parse(time.RFC3339, "2022-02-15T13:00:00Z")
if err != nil {
    // handle error
}

// use the parsed time value
```

## Panic Behavior

If the input string value is not a valid RFC3339 formatted date time string, this function will panic with an error.

## Benefits of Using this Function

Using this function can help to simplify your code by avoiding the need to handle errors returned from time.Parse. It can also help to prevent potential bugs caused by invalid input data.

## Limitations

This function is intended for use with RFC3339 formatted date time strings only. If you need to parse a different format, you may need to use a different function or library.
