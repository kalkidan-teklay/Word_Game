let wordList = ["WORD", "APPLE", "ORANGE", "BANANA"]; // List of words
let currentWordIndex = 0;
let score = 0; // Player's score
let word = wordList[currentWordIndex];

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

// Start game
window.onload = () => startGame();

function startGame() {
    displayWord(word);
}

// Display the scrambled word
function displayWord(word) {
    const shuffledWord = shuffleString(word);
    const paragraph = document.getElementById("generated_text");

    if (paragraph) {
        const wrappedText = shuffledWord.split('').map(letter => {
            return `<span class="letters">${letter}</span>`;
        }).join('');
        paragraph.innerHTML = wrappedText;
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

// Check the player's guess (NO SERVER REQUEST)
function checkAnswer() {
    const inputBoxes = document.querySelectorAll('.input-box');
    let userGuess = '';

    inputBoxes.forEach(input => {
        userGuess += input.value.trim().toUpperCase();
    });

    const resultMessage = document.getElementById("result_message");
    resultMessage.style.visibility = "visible";
    resultMessage.style.animation = "fadeIn 1s ease, zoom-in-zoom-out 1s ease infinite"; 

    setTimeout(() => {
        resultMessage.style.visibility = "hidden";
    }, 2000);

    if (userGuess === word) {
        resultMessage.textContent = "Correct!";
        resultMessage.style.color = "rgb(156, 236, 35)";
        document.getElementById("correct_sound").play(); 
        
        score++; // Increase score
        updatePlayerScore();

        // Move to the next word
        currentWordIndex++;
        if (currentWordIndex < wordList.length) {
            word = wordList[currentWordIndex];
            displayWord(word);
        } else {
            resultMessage.textContent = "Game Over! You completed all words!";
            setTimeout(() => {
                window.location.reload();
            }, 3000);
        }
    } else {
        resultMessage.textContent = "Incorrect!";
        resultMessage.style.color = "red";
        document.getElementById("wrong-sound").play();
    }
}

// Update the score display
function updatePlayerScore() {
    const playerListContainer = document.getElementById("player-list");
    playerListContainer.innerHTML = `Player: ${score}`;
}

// Event listeners
document.addEventListener("DOMContentLoaded", function () {
    const submitButton = document.getElementById("submit_button");
    if (submitButton) {
        submitButton.addEventListener("click", checkAnswer);
    }
});
