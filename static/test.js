  // Вопросы для теста
  const questions = [
    {
        question: "Какой язык используется для стилизации веб-страниц?",
        options: ["HTML", "CSS", "JavaScript", "PHP"],
        answer: 1
    },
    {
        question: "Что такое DOM в контексте веб-разработки?",
        options: [
            "Объектная модель документа", 
            "Тип базы данных", 
            "Язык программирования", 
            "Серверный фреймворк"
        ],
        answer: 0
    },
    {
        question: "Какой метод HTTP используется для запроса данных с сервера?",
        options: ["POST", "PUT", "GET", "DELETE"],
        answer: 2
    },
    {
        question: "Что из перечисленного НЕ является фреймворком JavaScript?",
        options: ["React", "Vue", "Angular", "Django"],
        answer: 3
    },
  
    {
        question: "Что такое API?",
        options: [
            "Автоматическая обработка информации", 
            "Интерфейс программирования приложений", 
            "Асинхронное программирование", 
            "Протокол передачи данных"
        ],
        answer: 1
    },
    {
        question: "Какой селектор CSS выбирает элемент по его идентификатору?",
        options: [".class", "#id", "*", "tag"],
        answer: 1
    },
    {
        question: "Что из перечисленного НЕ является типом данных в JavaScript?",
        options: ["String", "Boolean", "Float", "Object"],
        answer: 2
    },
    {
        question: "Какой метод добавляет новый элемент в конец массива в JavaScript?",
        options: ["push()", "pop()", "shift()", "unshift()"],
        answer: 0
    },
    {
        question: "Что такое Git?",
        options: [
            "Язык программирования", 
            "Система контроля версий", 
            "Фреймворк для тестирования", 
            "Графический редактор"
        ],
        answer: 1
    }
];

// Элементы DOM
const questionText = document.getElementById('question-text');
const optionsContainer = document.getElementById('options-container');
const currentQuestionEl = document.getElementById('current-question');
const totalQuestionsEl = document.getElementById('total-questions');
const totalQuestionsEl2 = document.getElementById('total-questions2');
const correctCountEl = document.getElementById('correct-count');
const progressFill = document.getElementById('progress-fill');
const nextBtn = document.getElementById('next-btn');
const prevBtn = document.getElementById('prev-btn');
const finishBtn = document.getElementById('finish-btn');
const resultsContainer = document.getElementById('results-container');
const finalScore = document.getElementById('final-score');
const scoreText = document.getElementById('score-text');
const resultDetails = document.getElementById('result-details');
const restartBtn = document.getElementById('restart-btn');
const timerEl = document.getElementById('timer');
const qNumber = document.getElementById('q-number');

// Переменные состояния
let currentQuestion = 0;
let userAnswers = Array(questions.length).fill(null);
let correctAnswers = 0;
let timer;
const totalTime = 10 * 60; // 10 минут в секундах
let timeLeft = totalTime;

// Инициализация теста
function initTest() {
    totalQuestionsEl.textContent = questions.length;
    totalQuestionsEl2.textContent = questions.length;
    qNumber.textContent = currentQuestion + 1;
    updateProgressBar();
    updateQuestion();
    startTimer();
}

// Обновление вопроса
function updateQuestion() {
    const question = questions[currentQuestion];
    questionText.textContent = question.question;
    qNumber.textContent = currentQuestion + 1;
    currentQuestionEl.textContent = currentQuestion + 1;
    
    // Очистка контейнера с вариантами
    optionsContainer.innerHTML = '';
    
    // Добавление вариантов ответа
    question.options.forEach((option, index) => {
        const optionEl = document.createElement('div');
        optionEl.className = 'option';
        if (userAnswers[currentQuestion] === index) {
            optionEl.classList.add('selected');
        }
        
        optionEl.innerHTML = `
            <div class="option-label"></div>
            <div class="option-text">${option}</div>
        `;
        
        optionEl.addEventListener('click', () => {
            // Снятие выделения со всех вариантов
            document.querySelectorAll('.option').forEach(opt => {
                opt.classList.remove('selected');
            });
            
            // Выделение выбранного варианта
            optionEl.classList.add('selected');
            userAnswers[currentQuestion] = index;
            
            // Проверка ответа
            if (index === question.answer) {
                correctAnswers++;
                correctCountEl.textContent = correctAnswers;
            } else if (userAnswers[currentQuestion] !== null && userAnswers[currentQuestion] !== question.answer) {
                correctCountEl.textContent = correctAnswers;
            }
        });
        
        optionsContainer.appendChild(optionEl);
    });
    
    // Обновление состояния кнопок
    prevBtn.disabled = currentQuestion === 0;
    nextBtn.disabled = currentQuestion === questions.length - 1;
}

// Обновление прогресс бара
function updateProgressBar() {
    const progress = ((currentQuestion + 1) / questions.length) * 100;
    progressFill.style.width = `${progress}%`;
}

// Переключение вопросов
nextBtn.addEventListener('click', () => {
    if (currentQuestion < questions.length - 1) {
        currentQuestion++;
        updateQuestion();
        updateProgressBar();
    }
});

prevBtn.addEventListener('click', () => {
    if (currentQuestion > 0) {
        currentQuestion--;
        updateQuestion();
        updateProgressBar();
    }
});

// Завершение теста
finishBtn.addEventListener('click', () => {
    showResults();
    clearInterval(timer);
});

// Перезапуск теста
restartBtn.addEventListener('click', () => {
    currentQuestion = 0;
    userAnswers = Array(questions.length).fill(null);
    correctAnswers = 0;
    correctCountEl.textContent = '0';
    timeLeft = totalTime;
    updateTimerDisplay();
    
    document.querySelector('.question-container').style.display = 'block';
    document.querySelector('.controls').style.display = 'flex';
    resultsContainer.style.display = 'none';
    
    updateQuestion();
    updateProgressBar();
    startTimer();
});

// Показ результатов
function showResults() {
    // Расчет результата
    const score = Math.round((correctAnswers / questions.length) * 100);
    finalScore.textContent = `${score}%`;
    
    // Текстовое описание результата
    let text = '';
    if (score >= 90) {
        text = 'Отличный результат! Вы настоящий эксперт!';
    } else if (score >= 70) {
        text = 'Хороший результат! Вы хорошо разбираетесь в теме.';
    } else if (score >= 50) {
        text = 'Неплохо! Но есть над чем поработать.';
    } else {
        text = 'Попробуйте еще раз! Вы можете лучше!';
    }
    scoreText.textContent = text;
    
    resultDetails.textContent = 
        `Вы ответили правильно на ${correctAnswers} из ${questions.length} вопросов`;
    
    // Скрыть вопрос и показать результаты
    document.querySelector('.question-container').style.display = 'none';
    document.querySelector('.controls').style.display = 'none';
    resultsContainer.style.display = 'block';
}

// Таймер
function startTimer() {
    clearInterval(timer);
    timer = setInterval(() => {
        timeLeft--;
        updateTimerDisplay();
        
        if (timeLeft <= 0) {
            clearInterval(timer);
            showResults();
        }
    }, 1000);
}

function updateTimerDisplay() {
    const minutes = Math.floor(timeLeft / 60);
    const seconds = timeLeft % 60;
    timerEl.textContent = `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
    
    // Изменение цвета при низком времени
    if (timeLeft < 60) {
        timerEl.style.color = '#ff5252';
        timerEl.style.animation = 'pulse 1s infinite';
    } else {
        timerEl.style.color = '';
        timerEl.style.animation = '';
    }
}

// Запуск теста при загрузке страницы
document.addEventListener('DOMContentLoaded', initTest);