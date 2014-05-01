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
     // connect to central server to obtain address for game server
     var central_ws = new WebSocket("ws://unix5.andrew.cmu.edu:8085/");
     
     central_ws.onopen = function() {
         console.log("connection to central server established");
     }
     
     central_ws.onmessage = function(event) {
         //central server will send a hostport for a game server upon
         //connection
         var gs_port = event.data;
         console.log("Being redirected to game server at port " + gs_port)
         var game_ws = new WebSocket("ws://unix5.andrew.cmu.edu:" + gs_port + "/");
         
         game_ws.onopen = function() {
             console.log("connected to game server")
             $('#start').hide();
             MakeTable(6,6);
             $('body').on("keydown", function (event) {
                     switch (event.which) {
                     case 37: //left arrow
                         console.log("pressed left");
                         game_ws.send("Left");
                         break;
                     case 39: //right arrow
                         console.log("pressed right");
                         game_ws.send("Right");
                         break;
                     case 40: //down arrow
                         game_ws.send("Down");
                         break;
                     case 80: //p key
                         game_ws.send("Pause");
                         break;
                     }
                 });
         }
         
         game_ws.onmessage = function (event) { 
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
                         $(id).css("background-color", "#FF6600");
                         break;
                     case 3: 
                         $(id).css("background-color", "#7AC861");
                         break;
                     }
                 }
             }
         }

         game_ws.onclose = function() {
             console.log("Connection to game server closed")
         }
     }
     
     central_ws.onclose = function() {
         console.log("Connection to central server closed")
     }
  }
  else
  {
     // The browser doesn't support WebSocket
     console.log("WebSocket NOT supported by your Browser!");
  }
}
