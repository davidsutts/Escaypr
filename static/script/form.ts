// Global variables.
var loading = false;

// Check for changes on form to update submission button.
document.addEventListener("DOMContentLoaded", () => {
	document.getElementById("pword")?.addEventListener("change", toggleSubmit);
	document.getElementById("uname")?.addEventListener("change", toggleSubmit);
})

// Submit form when enter is pressed.
document.addEventListener("keyup", function(event: KeyboardEvent) {
	var button = document.querySelector<HTMLElement>("#submit");
	if (event.code == "Enter") {
		if (checkForm()) {
			postForm();
			button?.setAttribute("disabled", "true");
		}
	} else {
		return;
	}
});

// Check the form to ensure the username and password are not blank
function checkForm() : boolean {
	let inputs = document.querySelectorAll("input");
	for (let i = 0; i < 2; i++) {
		if (loading || inputs[i].value == "") {
			return false;
		}
	}
	return true;
}

// Toggle the submit button depending on the checkForm return.
function toggleSubmit() {
	var button = document.querySelector<HTMLElement>("#submit");
	if (checkForm() == false){
		button?.setAttribute("disabled", "true");
	} else {
		button?.removeAttribute("disabled");
	}
}

// Post the inputs to the form.
function postForm() {	
	const msg = document.querySelector("p")!;
	const form = document.querySelector('form')!;
	const url =  "/login/form";

	const formData = new FormData(form);

	let xhr = new XMLHttpRequest;
	xhr.open("POST", url);

	// Load response and redirect or fail attempt.
	xhr.onloadstart = () => {loading = true;}
	xhr.onloadend = () => {
		msg.style.display = "block";
		if (xhr.status == 200) {
			msg.innerText = "Login Successful";
			loading = false;
			window.location.replace("/index")	// Send logged in user to index.
		} else {
			msg.innerText = "Invalid usernamne or password";
			setTimeout(() => {
				loading = false;
			}, 1000);	// Make user wait before trying again.
		}
	}

	// Send login request.
	xhr.send(formData);

}