Find n (default 1) term colors closest to the color specified via command-line argument.
Outputs color term code, hex code, name, and distance.

Examples:

Outputs the closest color
```
go run closest-term-colors.go '#3352ce'
```

Outputs 5 closest colors in increasing distance order.
```
go run closest-term-colors.go '#3352ce' 5
```
