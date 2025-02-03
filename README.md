# README

## Description

This project is an interactive application that displays a rotating sphere with zoom and drag functionality. The sphere is made of triangles that animate and change color based on time and level of detail (LOD).

## Requirements

- Install [Go](https://golang.org/dl/)
- Install Ebiten:
  ```
  go get -u github.com/hajimehoshi/ebiten/v2
  ```

## How to Run

1. Clone the repository or copy the code into `main.go`.
2. Navigate to the project directory.
3. Run:
   ```
   go run main.go
   ```

## Controls

- **Left/Right Arrow**: Change rotation direction.
- **Mouse Wheel**: Zoom in/out.
- **Left Click**: Drag the window if the cursor is over the sphere.

## Features

- **Zoom**: Adjusts the sphere's level of detail.
- **Animation**: Sphere vertices animate over time.
- **Colors**: Triangle colors smoothly transition.
