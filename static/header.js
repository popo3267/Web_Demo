function loadHTML(id, url) {
    fetch(url)
        .then(response => response.text())
        .then(data => {
            document.getElementById(id).innerHTML = data;

            if (id === 'header') {
                executeAuthNavFunctionality();
            }
        })
}

document.addEventListener("DOMContentLoaded", function() {
    loadHTML("header", "/Member/header");
    loadHTML("footer", "/Member/footer");
});

function executeAuthNavFunctionality() {
    fetch('/Member/status')
        .then(response => response.json())
        .then(data => {
            const authNav = document.getElementById('auth-nav');

            if (authNav) {
                if (data.logged_in) {
                    authNav.innerHTML = `
                        <li class="nav-item">
                            <span class="nav-link link-light px-2">${data.member}</span>
                        </li>
                        <li class="nav-item">
                            <a href="/Member/Logout" class="nav-link link-light px-2">登出</a>
                        </li>
                    `;
                } else {
                    authNav.innerHTML = `
                        <li class="nav-item">
                            <a href="/Member/Register" class="nav-link link-light px-2">註冊</a>
                        </li>
                        <li class="nav-item">
                            <a href="/Member/Login" class="nav-link link-light px-2">登入</a>
                        </li>
                    `;
                }
            } else {
                
            }
        })
       
}
