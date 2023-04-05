document.addEventListener("keydown", function (event) {
    var button = document.querySelector("#submit");
    if (checkForm() == false) {
        button === null || button === void 0 ? void 0 : button.setAttribute("disabled", "true");
    }
    else {
        button === null || button === void 0 ? void 0 : button.removeAttribute("disabled");
        if (event.code == "Enter") {
            postForm();
        }
    }
});
function checkForm() {
    var inputs = document.querySelectorAll("input");
    for (var i = 0; i < 2; i++) {
        if (inputs[i].value == "") {
            return false;
        }
    }
    return true;
}
function postForm() {
    var message = document.querySelector('p.err');
    if (!checkForm()) {
        if (message === null || message === void 0 ? void 0 : message.style) {
            message.style.display = "block";
            message.innerText = "Invalid username or password";
        }
        return;
    }
    else {
        if (message === null || message === void 0 ? void 0 : message.style) {
            message.style.display = "none";
        }
    }
    var form = document.querySelector('form');
    var url = "/login/form";
    var formData = new FormData(form);
    var xhr = new XMLHttpRequest;
    xhr.open("POST", url);
    xhr.send(formData);
}
