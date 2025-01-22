let word = '';
let result = 0; 

function createInputBoxes(word) {
    const input_container = document.getElementById("input_boxes");
    input_container.innerHTML = ''; 
    for (let char of word) {
        const input = document.createElement('input');
        input.type = 'text';
        input.maxLength = 1;
        input.className = 'input-box';
        input_container.appendChild(input);
    }
}

async function fetchRandomWord() {
    try {
        const response = await fetch('https://random-word-api.herokuapp.com/word?number=1');
        const data = await response.json();
        
        if (data.length > 0) {
            word = data[0]; // Assign fetched word to global variable
            console.log(word);
            const shuffledWord = shuffleString(word);
            const paragraph = document.getElementById("generated_text");
            
            createInputBoxes(word);

            const wrappedText = shuffledWord.split('').map(letter => {
                return `<span class="letters">${letter}</span>`;
            }).join('');

            paragraph.innerHTML = wrappedText; 
        }
    } catch (error) {
        console.error('Error fetching random word:', error);
    }
}

function shuffleString(str) {
    const arr = str.split(''); 
    for (let i = arr.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1)); 
        [arr[i], arr[j]] = [arr[j], arr[i]];
    }
    return arr.join(''); 
}

function checkAnswer() {
    const inputBoxes = document.querySelectorAll('.input-box');
    const score = document.getElementById("value");
    
    let userGuess = '';

    inputBoxes.forEach(input => {
        userGuess += input.value.trim(); 
    });

    const resultMessage = document.getElementById("result_message");
    console.log("word: ", word);
    console.log("userGuess: ", userGuess);
    if (userGuess.toLowerCase() === word.toLowerCase()) {
        result+= 1
        resultMessage.textContent = "Correct! You guessed the word!";
        score.innerHTML = result;
        fetchRandomWord();
    } else {
        resultMessage.textContent = "Incorrect! Try again.";
    }
}

document.getElementById("submit_button").addEventListener("click", checkAnswer);
window.onload = fetchRandomWord;
