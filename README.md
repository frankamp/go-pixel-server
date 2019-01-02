Pixel Display Server
====================

This program launches a window and a localhost daemon that listens on 8080 for a json packet.

If you supply a scene packet like

    {
        "b": {"e":[{"r":[50,50,500,51]}]},
        "f":[
            {"e":[{"r":[100,200,300,400]}]},
            {"e":[{"c":{"n":"color","v":"1,0,0"}},{"r":[150,250,350,450]}]}
        ]
    }

It will draw a baseframe with a single pixel high, wide rectangle, and add two frames to the scene. The first of the two will draw a black square. The second will draw a red square 50 pixels offset diagonally.

Using left right keys you can move between frames. Use mouse scroll to zoom. Use Z to to return to a default camera location. Use WASD to pan.

Other command elements like color `{"c":{"n":"color","v":"1,0,0"}}` are:
* thickness, 0 is filled, above that will be converted to a float and used as faiface/pixel uses thickness
* color, RGB three floats between 0-1, default color is black
* shape: rectangle (default), or circle
** if you change to circle, then a region element {"r":[150,250,350,450]} is interpreted as circle centered at 150,250, with ellipse radii 350,450. To make a true circle, set the third and fourth to the same radius.

Example scenes I pushed from client programs are like:

 - `raw.githubusercontent.com/frankamp/go-pixel-server/master/final.gif?sanitize=true&raw=true`