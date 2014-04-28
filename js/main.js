//generates an HTML table with the given height and width
function MakeTable (height, width) {
    var displayStr = "";
    for (var i = 0; i < height; i++) {
        displayStr = displayStr.concat(' <tr> ')
        for (var j = 0; j < width; j++) {
            displayStr = displayStr.concat('<td id = "',
                                           i.toString(),
                                           '_', 
                                           j.toString(),
                                           '"></td>');
        };
        displayStr = displayStr.concat(' </tr> ');
    };
    $('#Table').append(displayStr);    
};

function PosToId(pos) {
    return ('#').concat(pos.Row.toString(), "_", pos.Col.toString())
}


function Start(){
  if ("WebSocket" in window)
  {
     console.log("WebSocket is supported by your Browser!");
     // Let us open a web socket
     var ws = new WebSocket("ws://localhost:8080/");

     ws.onopen = function() {
         $('#start').hide();
         MakeTable(4,4);
         $('body').on("keydown", function (event) {
                 switch (event.which) {
                 case 37: //left arrow
                     console.log("pressed left");
                     ws.send("Left");
                     break;
                 case 39: //right arrow
                     console.log("pressed right");
                     ws.send("Right");
                     break;
                 case 40: //down arrow
                     ws.send("Down");
                     break;
                 case 80: //p key
                     ws.send("Pause");
                     break;
                 }
             });
     }

     ws.onmessage = function (event) { 
        var received_msg = event.data;

        if (received_msg === "GameOver") {
            $('#Table').fadeTo(1,.3);
            $('body').off();
            $('#dialogue').css("display","block");
            $('#dialogue').html('Game Over');
            
        } else {
            var updates = jQuery.parseJSON(received_msg);

            for (var i = 0; i < updates.length; i++){
                var id = PosToId(updates[i].Pos);
                var val = updates[i].Value;
                switch (val) {
                case 0: 
                    $(id).css("background-color", "white");
                    break;
                case 1: 
                    $(id).css("background-color", "#3399FF");
                    break;
                case 2: 
                    $(id).css("background-color", "#7AC861");
                break;
                }
            }
        }
     }

     ws.onclose = function()
     { 
        // websocket is closed.
        console.log("Connection is closed..."); 
     };
  }
  else
  {
     // The browser doesn't support WebSocket
     console.log("WebSocket NOT supported by your Browser!");
  }
}
