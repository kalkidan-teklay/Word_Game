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


// Start the game or fetch the first word
async function startGame() {
    const userId = localStorage.getItem("userId");
    if (!userId) {
        alert("Please log in first.");
        return;
    }

    try {
        const response = await fetch('http://localhost:8080/start', {
            method: 'GET',
            headers: { 'Content-Type': 'application/json' },
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

    const wrappedText = shuffledWord.split('').map(letter => {
        return `<span class="letters">${letter}</span>`;
    }).join('');

    paragraph.innerHTML = wrappedText;
    createInputBoxes(word);
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
        const resultMessage = document.getElementById("result_message");
        resultMessage.style.visibility = "visible";
        resultMessage.style.animation = "fadeIn 1s ease, zoom-in-zoom-out 1s ease infinite"; 
        setTimeout(() => {
            resultMessage.style.visibility = "hidden";
        }, 2000);

        if (data.correct) {
            resultMessage.textContent = "Correct!";
            const correctSound = document.getElementById("correct_sound");
                correctSound.play(); 
            word = data.new_word;
            displayWord(word); // Display the next word
        } else {
            
            resultMessage.textContent = "Incorrect!";
            const IncorrectSound = document.getElementById("wrong-sound");
                IncorrectSound.play(); 
        }

        // Update the scores
        updateScores(data.scores);

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
function updateScores(scores) {
    const scoreContainer = document.getElementById("score");
    if (scores && Array.isArray(scores)) {
        scoreContainer.innerHTML = `Scores: ${scores.map(score => `${score.name}: ${score.points}`).join(' | ')}`;
    } else {
        scoreContainer.innerHTML = "Scores: N/A"; // Handle undefined or invalid scores
    }
}

// Event listeners
document.addEventListener("DOMContentLoaded", function () {
    const submitButton = document.getElementById("submit_button");
    if (submitButton) {
        submitButton.addEventListener("click", checkAnswer);
    }

    
    window.onload = startGame;
});