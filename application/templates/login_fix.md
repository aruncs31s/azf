// package templates

// templ LoginPage(theError string) {
// <!DOCTYPE html>
// <html lang="en">
// <head>
// <meta charset="UTF-8"/>
// <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
// <title>Admin Login - Permission Management</title>
// <script src="https://cdn.tailwindcss.com"></script>
// <script>
// tailwind.config = {
// darkMode: 'class',
// }
// </script>
// <style>
// body {
// background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
// min-height: 100vh;
// }
// .dark body {
// background: linear-gradient(135deg, #1e1b4b 0%, #2e1065 100%);
// }
// .login-card {
// background: rgba(255, 255, 255, 0.95);
// backdrop-filter: blur(10px);
// }
// .dark .login-card {
// background: rgba(31, 41, 55, 0.95);
// backdrop-filter: blur(10px);
// }
// .login-btn:hover {
// transform: translateY(-2px);
// box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
// }
// .input-focus:focus {
// border-color: #667eea;
// box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
// }
// .dark .input-focus:focus {
// border-color: #818cf8;
// box-shadow: 0 0 0 3px rgba(129, 140, 248, 0.1);
// }
// </style>
// <script>
// function storeJWTToken(token) {
// localStorage.setItem('jwt_token', token);
// }

// function getJWTToken() {
// return localStorage.getItem('jwt_token');
// }

// function clearJWTToken() {
// localStorage.removeItem('jwt_token');
// }

// // Simple token validation - just check if token exists
// // DONT make HTTP requests to validate to avoid redirect loops
// function isTokenValid() {
// const token = getJWTToken();
// return token && typeof token === 'string' && token.length > 0;
// }

// async function handleLoginJSON(event) {
// event.preventDefault();
// const username = document.getElementById('username').value;
// const password = document.getElementById('password').value;

// if (!username || !password) {
// alert('Please enter both username and password');
// return;
// }

// try {
// const response = await fetch('/admin-ui/login/json', {
// method: 'POST',
// headers: {
// 'Content-Type': 'application/json',
// },
// body: JSON.stringify({ username, password }),
// });

// const data = await response.json();

// if (response.ok && data.success && data.jwt) {
// storeJWTToken(data.jwt);
// window.location.href = '/admin-ui';
// } else {
// clearJWTToken();
// alert('Login failed: ' + (data.message || 'Unknown error'));
// }
// } catch (error) {
// console.error('Login error:', error);
// clearJWTToken();
// alert('An error occurred during login');
// }
// }

// function initializeDarkMode() {
// const isDarkMode = localStorage.getItem('darkMode') === 'true';
// const htmlElement = document.documentElement;
// if (isDarkMode) {
// htmlElement.classList.add('dark');
// } else {
// htmlElement.classList.remove('dark');
// }
// }

// window.addEventListener('load', function() {
// const token = getJWTToken();
// const currentPath = window.location.pathname;
// if (token && (currentPath === '/admin-ui/login' || currentPath === '/login')) {
// if (isTokenValid()) {
// window.location.href = '/admin-ui';
// } else {
// clearJWTToken();
// }
// }
// });

// initializeDarkMode();
// document.addEventListener('DOMContentLoaded', initializeDarkMode);
// </script>
// </head>
// <body class="flex items-center justify-center dark:bg-gray-950">
// <div class="w-full max-w-md">
// <div class="text-center mb-8">
// <h1 class="text-4xl font-bold text-white dark:text-gray-100 mb-2">Admin Panel</h1>
// <p class="text-gray-200 dark:text-gray-400">Permission Management System</p>
// </div>
// <div class="login-card rounded-2xl shadow-2xl p-8">
// <div class="mb-8">
// <h2 class="text-2xl font-bold text-gray-800 dark:text-gray-100 mb-2">Welcome Back</h2>
// <p class="text-gray-600 dark:text-gray-400">Sign in to your admin account</p>
// </div>
// if theError != "" {
// <div class="mb-6 p-4 bg-red-50 dark:bg-red-900 border-l-4 border-red-500 rounded">
// <p class="text-red-700 dark:text-red-200 text-sm">{ theError }</p>
// </div>
// }
// <form onsubmit="handleLoginJSON(event)" class="space-y-5">
// <div>
// <label for="username" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
// Username
// </label>
// <input
// type="text"
// id="username"
// name="username"
// placeholder="Enter your username"
// required
// class="input-focus w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none transition duration-300 dark:bg-gray-800 dark:text-gray-100"
// />
// </div>
// <div>
// <label for="password" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
// Password
// </label>
// <input
// type="password"
// id="password"
// name="password"
// placeholder="Enter your password"
// required
// class="input-focus w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none transition duration-300 dark:bg-gray-800 dark:text-gray-100"
// />
// </div>
// <button
// type="submit"
// class="login-btn w-full bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white font-bold py-3 px-4 rounded-lg transition duration-300 ease-in-out mt-6"
// >
// Sign In
// </button>
// </form>
// <div class="my-6 flex items-center">
// <div class="flex-1 border-t border-gray-300 dark:border-gray-600"></div>
// <span class="px-2 text-sm text-gray-500 dark:text-gray-400">Admin Only</span>
// <div class="flex-1 border-t border-gray-300 dark:border-gray-600"></div>
// </div>
// </div>
// <div class="mt-8 text-center text-gray-300 dark:text-gray-500 text-sm">
// <p>© 2024 Admin Panel. All rights reserved.</p>
// </div>
// </div>
// </body>
// </html>
// }
