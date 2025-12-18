// Get elements
const emailInput = document.getElementById('email');
const passwordInput = document.getElementById('password');
const emailError = document.getElementById('emailError');
const passwordError = document.getElementById('passwordError');
const togglePassword = document.getElementById('togglePassword');
const loginForm = document.getElementById('loginForm');
const btnLogin = document.getElementById('btnLogin');

// Toggle password visibility
togglePassword.addEventListener('click', function() {
    const type = passwordInput.getAttribute('type') === 'password' ? 'text' : 'password';
    passwordInput.setAttribute('type', type);
    
    // Toggle icon
    this.classList.toggle('fa-eye');
    this.classList.toggle('fa-eye-slash');
});

// Email validation (live)
emailInput.addEventListener('input', function() {
    validateEmail(this.value, true);
});

emailInput.addEventListener('blur', function() {
    validateEmail(this.value, false);
});

// Password validation (live)
passwordInput.addEventListener('input', function() {
    validatePassword(this.value, true);
});

passwordInput.addEventListener('blur', function() {
    validatePassword(this.value, false);
});

// Validate Email Function
function validateEmail(email, isTyping = false) {
    email = email.trim();
    
    // Clear error saat typing
    if (isTyping && email.length < 5) {
        emailError.textContent = '';
        emailInput.classList.remove('error', 'success');
        return false;
    }
    
    if (email === '') {
        emailError.textContent = 'Email tidak boleh kosong';
        emailInput.classList.add('error');
        emailInput.classList.remove('success');
        return false;
    }
    
    const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
    if (!emailRegex.test(email)) {
        emailError.textContent = 'Format email tidak valid';
        emailInput.classList.add('error');
        emailInput.classList.remove('success');
        return false;
    }
    
    if (email.length < 5) {
        emailError.textContent = 'Email terlalu pendek';
        emailInput.classList.add('error');
        emailInput.classList.remove('success');
        return false;
    }
    
    if (email.length > 100) {
        emailError.textContent = 'Email terlalu panjang (maksimal 100 karakter)';
        emailInput.classList.add('error');
        emailInput.classList.remove('success');
        return false;
    }
    
    // Valid
    emailError.textContent = '';
    emailInput.classList.remove('error');
    emailInput.classList.add('success');
    return true;
}

// Validate Password Function
function validatePassword(password, isTyping = false) {
    // Clear error saat typing
    if (isTyping && password.length < 3) {
        passwordError.textContent = '';
        passwordInput.classList.remove('error', 'success');
        return false;
    }
    
    if (password === '') {
        passwordError.textContent = 'Password tidak boleh kosong';
        passwordInput.classList.add('error');
        passwordInput.classList.remove('success');
        return false;
    }
    
    if (password.length < 3) {
        passwordError.textContent = 'Password minimal 3 karakter';
        passwordInput.classList.add('error');
        passwordInput.classList.remove('success');
        return false;
    }
    
    if (password.length > 50) {
        passwordError.textContent = 'Password maksimal 50 karakter';
        passwordInput.classList.add('error');
        passwordInput.classList.remove('success');
        return false;
    }
    
    // Valid
    passwordError.textContent = '';
    passwordInput.classList.remove('error');
    passwordInput.classList.add('success');
    return true;
}

// Form Submit
loginForm.addEventListener('submit', async function(e) {
    e.preventDefault();
    
    const email = emailInput.value.trim();
    const password = passwordInput.value;
    
    // Validate before submit
    const isEmailValid = validateEmail(email, false);
    const isPasswordValid = validatePassword(password, false);
    
    if (!isEmailValid || !isPasswordValid) {
        return;
    }
    
    // Disable button saat loading
    btnLogin.disabled = true;
    btnLogin.textContent = 'Loading...';
    
    try {
        const response = await fetch('/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                email: email.toLowerCase(),
                password: password
            })
        });
        
        const data = await response.json();
        
        if (data.success) {
            // Success alert
            await Swal.fire({
                icon: 'success',
                title: 'Login Berhasil!',
                text: `Selamat datang, ${data.user.nama}`,
                timer: 2000,
                showConfirmButton: false
            });
            
            // Redirect ke dashboard
            window.location.href = '/employees';
        } else {
            // Error handling berdasarkan field
            if (data.field === 'email') {
                emailError.textContent = data.message;
                emailInput.classList.add('error');
                emailInput.focus();
            } else if (data.field === 'password') {
                passwordError.textContent = data.message;
                passwordInput.classList.add('error');
                passwordInput.value = ''; // Reset password
                passwordInput.focus();
            } else {
                // General error
                Swal.fire({
                    icon: 'error',
                    title: 'Login Gagal!',
                    text: data.message,
                    confirmButtonColor: '#667eea'
                });
                
                // Reset hanya password
                passwordInput.value = '';
                passwordInput.classList.remove('error', 'success');
                passwordError.textContent = '';
            }
        }
    } catch (error) {
        Swal.fire({
            icon: 'error',
            title: 'Oops...',
            text: 'Terjadi kesalahan pada server!',
            confirmButtonColor: '#667eea'
        });
        
        // Reset password
        passwordInput.value = '';
    } finally {
        // Enable button kembali
        btnLogin.disabled = false;
        btnLogin.textContent = 'Login';
    }
});
