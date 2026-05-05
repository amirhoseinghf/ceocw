// signup.js
document.addEventListener('DOMContentLoaded', function() {
    const form = document.querySelector('.auth-form');
    if (!form) return;

    form.addEventListener('submit', function(e) {
        e.preventDefault();

        const firstName = document.getElementById('first_name').value.trim();
        const lastName = document.getElementById('last_name').value.trim();
        const email = document.getElementById('email').value.trim();
        const password = document.getElementById('password').value;

        // Name validations
        if (!firstName) {
            showToast('لطفاً نام خود را وارد کنید', false);
            return;
        }
        if (!isPersianText(firstName)) {
            showToast('نام باید به فارسی وارد شود', false);
            return;
        }
        if (firstName.length > 50) {
            showToast('نام نمی‌تواند بیش از ۵۰ کاراکتر باشد', false);
            return;
        }
        if (!lastName) {
            showToast('لطفاً نام خانوادگی خود را وارد کنید', false);
            return;
        }
        if (lastName.length > 50) {
            showToast('نام خانوادگی نمی‌تواند بیش از ۵۰ کاراکتر باشد', false);
            return;
        }
        if (!isPersianText(lastName)) {
            showToast('نام خانوادگی باید به فارسی وارد شود', false);
            return;
        }

        // Email validation
        if (!email) {
            showToast('لطفاً ایمیل خود را وارد کنید', false);
            return;
        }
        if (email.length > 50) {
            showToast('ایمیا بیش از ۵۰ کاراکتر باشد', false);
            return;
        }
        if (!isValidEmail(email)) {
            showToast('لطفاً یک ایمیل معتبر وارد کنید', false);
            return;
        }

        // Password validation
        if (!password) {
            showToast('لطفاً رمز عبور را وارد کنید', false);
            return;
        }
        if (password.length < 8) {
            showToast('رمز عبور باید حداقل ۸ کاراکتر باشد', false);
            return;
        }

        if (password.length > 50) {
            showToast('رمز عبور نمی‌تواند بیش از ۵۰ کاراکتر باشد', false);
            return;
        }

        // If all valid, submit the form
        form.submit();
    });

    function isValidEmail(email) {
        const re = /^[^\s@]+@([^\s@.,]+\.)+[^\s@.,]{2,}$/;
        return re.test(email);
    }

    function isPersianText(str) {
        // Persian Unicode range: \u0600-\u06FF
        // Also allow space and the "آ" character (which is covered by the range)
        const persianRegex = /^[\u0600-\u06FF\s]+$/;
        return persianRegex.test(str);
    }
});