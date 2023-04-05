document.addEventListener("keydown", function(event: KeyboardEvent) {
	var button = document.querySelector<HTMLElement>("#submit");
	if (checkForm() == false){
		button?.setAttribute("disabled", "true");
	} else {
		button?.removeAttribute("disabled");
		if (event.code == "Enter") {
			postForm();
		}
	}
});

function checkForm() : boolean {
	let inputs = document.querySelectorAll("input");
	for (let i = 0; i < 2; i++) {
		if (inputs[i].value == "") {
			return false;
		}
	}
	return true;
}

function postForm() {
	var message = document.querySelector<HTMLElement>('p.err');

	if (!checkForm()) {
		if (message?.style) {
			message.style.display = "block";
			message.innerText = "Invalid username or password";
		}
		return;
	} else {
		if (message?.style) {
			message.style.display = "none";
		}
	}
	
	const form = document.querySelector('form')!;
	const url =  "/login/form";

	const formData = new FormData(form);

	let xhr = new XMLHttpRequest;
	xhr.open("POST", url);
	xhr.send(formData);

}