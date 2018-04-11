earcut
======

A library for efficiently transforming polygons into triangles in go.

[![Build Status](https://travis-ci.org/rclancey/go-earcut.svg?branch=master)](https://travis-ci.org/rclancey/go-earcut)
[![Go Report Card](https://goreportcard.com/badge/github.com/rclancey/go-earcut)](https://goreportcard.com/report/github.com/rclancey/go-earcut) 
[![Documentation](https://godoc.org/github.com/rclancey/go-earcut?status.svg)](http://godoc.org/github.com/rclancey/go-earcut)
[![Coverage Status](https://coveralls.io/repos/github/rclancey/go-earcut/badge.svg?branch=master)](https://coveralls.io/github/rclancey/go-earcut?branch=master)
[![GitHub issues](https://img.shields.io/github/issues/rclancey/go-earcut.svg)](https://github.com/rclancey/go-earcut/issues)
[![license](https://img.shields.io/github/license/rclancey/go-earcut.svg?maxAge=2592000)](https://github.com/rclancey/go-earcut/LICENSE)
[![Release](https://img.shields.io/github/release/rclancey/go-earcut.svg?label=Release)](https://github.com/rclancey/go-earcut/releases)

About
-----

This is a direct port of [Mapbox's JavaScript earcut library](https://github.com/mapbox/earcut).

Usage
-----

    // flat array of all vertices in polygon, including holes
    verts := []float64{
        0.0, 0.0,
        1.0, 0.0,
        1.0, 1.0,
        0.0, 1.0,
        0.0, 0.0,
        0.25, 0.25,
        0.75, 0.25,
        0.75, 0.75,
        0.25, 0.75,
        0.25, 0.25,
    }
    holes := []int{5} // *vertex* index of beginning of each hole
    dims := 2 // number of values per vertex
    indices, err := earcut.Earcut(verts, holes, dims)
    // indices is an array of integers (3 per triangle) referencing
    // the polygon vertexes that make up each triangle

Documentation
-------------

See [Mapbox](https://github.com/mapbox/earcut) and [GoDoc.org](https://godoc.org/github.com/rclancey/go-earcut) for complete documentation.

Author
------

Original implementation by [Mapbox](https://mapbox.com/).  Go port by [Ryan Clancey](https://github.com/rclancey)

License
-------

earcut is licensed under the [ISC license](https://en.wikipedia.org/wiki/ISC_license), described in the `LICENSE` file.
