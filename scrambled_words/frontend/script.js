let word = '';
let player_id = null; // Player's unique identifier

// Create input boxes for the current word
function createInputBoxes(word) {
    const inputContainer = document.getElementById("input_boxes");
    inputContainer.innerHTML = ''; 
    for (let char of word) {
        const input = document.createElement('input');
        input.type = 'text';
        input.maxLength = 1;
        input.className = 'input-box';
      

        input.addEventListener('input', (e) => {
            const nextInput = input.nextElementSibling;
            if (nextInput && nextInput.tagName === 'INPUT') {
                nextInput.focus();
            }
        });

        input.addEventListener('keydown', (e) => {
            if (e.key === 'Backspace' && input.value === '') {
                const previousInput = input.previousElementSibling;
                if (previousInput && previousInput.tagName === 'INPUT') {
                    previousInput.focus();
                }
            }
        });
        inputContainer.appendChild(input);
    }
}


window.onload = () => startGame(); 

async function startGame() {
    const userId = localStorage.getItem("userId");
    if (!userId) {
        alert("Please log in first.");
        return;
    }

    try {
        const payload = {
            player_id: userId, 
           
        };
        const response = await fetch('http://localhost:8080/start', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        });
        const data = await response.json();

        if (data.success) {
            player_id = data.player_id; // Assign the player ID
            word = data.word; // Get the word
            displayWord(word); // Show the word
        } else {
            alert(data.message || 'Error starting the game.');
        }
    } catch (error) {
        console.error('Error starting the game:', error);
    }
}




// Display the scrambled word
function displayWord(word) {
    const shuffledWord = shuffleString(word);
    const paragraph = document.getElementById("generated_text");

    // Ensure the element exists and is available
    if (paragraph) {
        const wrappedText = shuffledWord.split('').map(letter => {
            return `<span class="letters">${letter}</span>`;
        }).join('');

        // Safely set the innerHTML only when the element exists
        paragraph.innerHTML = wrappedText;

        // Proceed with creating input boxes for the word
        createInputBoxes(word);
    } else {
        console.error("The element with id 'generated_text' was not found.");
    }
}


    


// Shuffle a string (used for scrambling the word)
function shuffleString(str) {
    const arr = str.split(''); 
    for (let i = arr.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1)); 
        [arr[i], arr[j]] = [arr[j], arr[i]];
    }
    return arr.join(''); 
}

// Submit the player's guess
async function checkAnswer() {
    const userId = localStorage.getItem("userId");
    if (!userId) {
        alert("Please log in first.");
        return;
    }

    const inputBoxes = document.querySelectorAll('.input-box');
    let userGuess = '';

    inputBoxes.forEach(input => {
        userGuess += input.value.trim();
    });
    console.log("User ID:", userId); // Debugging
    console.log("User Guess:", userGuess); // Debugging

    try {
        const response = await fetch('http://localhost:8080/submit', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ player_id: userId, guess: userGuess }),
        });

        const data = await response.json();
        const playerScore = data.player ? data.player.score : "N/A"; 
        const resultMessage = document.getElementById("result_message");
        resultMessage.style.visibility = "visible";
        resultMessage.style.animation = "fadeIn 1s ease, zoom-in-zoom-out 1s ease infinite"; 
        setTimeout(() => {
            resultMessage.style.visibility = "hidden";
        }, 2000);

        if (data.correct) {
            resultMessage.textContent = "Correct!";
            resultMessage.style.color = "rgb(156, 236, 35);";
            const correctSound = document.getElementById("correct_sound");
                correctSound.play(); 
            word = data.new_word;
            
            
            displayWord(word); // Display the next word
        } else {
            
            resultMessage.textContent = "Incorrect!";
            resultMessage.style.color = "red";
            const IncorrectSound = document.getElementById("wrong-sound");
                IncorrectSound.play(); 
        }

        // Update the scores
        if (data.player) {
            updatePlayerList(data.player);
        }

        // Check for winner
        if (data.winner) {
            alert(`Game Over! ${data.winner} wins!`);
            window.location.reload(); // Reload the game
        }
    } catch (error) {
        console.error('Error checking answer:', error);
    }
}

// Update the score display
function updatePlayerList(players) {
    const playerListContainer = document.getElementById("player-list");
    
        playerListContainer.innerHTML = players
            .map(player => `${player.name}: ${player.score}`)
            .join("<br>");
        
    
}

// Event listeners
document.addEventListener("DOMContentLoaded", function () {
    const submitButton = document.getElementById("submit_button");
    if (submitButton) {
        submitButton.addEventListener("click", checkAnswer);
    }

    
    //window.onload = () => startGame(gameType); 
});


// Connect to WebSocket
const socket = new WebSocket('ws://localhost:8080/ws');

const name = localStorage.getItem("username");

// Send username to server after connection is open
socket.onopen = function () {
    if (name) {
        socket.send(JSON.stringify({
            type: "register",
            payload: {
                username: name,
            },
        }));
    } else {
        console.error("Username not found in local storage.");
    }
};

// Handle WebSocket messages
socket.addEventListener('message', (event) => {
    const message = JSON.parse(event.data);

    console.log('Received message:', message);
    const gameover = document.getElementById("game-over-container");
   
    if (message.type === "player_list") {
        console.log("Updated player list:", message.payload.players);
        updatePlayerList(message.payload.players);  // Function to update UI
    }
    if (message.type === 'game_over') {
        const winner = message.payload.winner;
        const currentUser = localStorage.getItem("username");
        gameover.style.visibility = "visible";
        if (currentUser === winner) {
            gameover.innerHTML = "🎉🎉YOU WON THE GAME!!"
            gameover.style.color = "white";
          
            gameover.style.fontSize = "24px"; 
            gameover.style.textAlign = "center"; 
            gameover.style.fontWeight = "bold"; 
            //alert("You won the game!");
        } else {
            gameover.innerHTML = "GAME OVER"
            
            alert(`${winner} won the game!`);
            gameover.style.color = "white";
            gameover.style.margin = "20px";
            gameover.style.fontSize = "24px"; 
            gameover.style.textAlign = "center"; 
            gameover.style.fontWeight = "bold"; 
        }
        setTimeout(() => {
            window.location.href = "./menu.html";   
        }, 4000);
        // Redirect to the game-over screen
        
    }
});

// Handle WebSocket errors
socket.addEventListener('error', (error) => {
    console.error('WebSocket error:', error);
});

// Handle WebSocket close
socket.addEventListener('close', () => {
    console.log('WebSocket connection closed');
});