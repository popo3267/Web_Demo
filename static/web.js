
  const togglePassword=document.querySelector('#togglePassword')
  const password=document.querySelector('#Password')
  togglePassword.addEventListener('click',function(){
    const type=password.getAttribute('type')=== 'password'?'text':'password'
    password.setAttribute('type', type)
    this.classList.toggle('bi-eye')
  })
  const toggleConfirmPassword=document.querySelector('#toggleConfirmPassword')
  const confirmpassword=document.querySelector('#ConfirmPassword')
  toggleConfirmPassword.addEventListener('click',function(){
    const type=confirmpassword.getAttribute('type')=== 'password'?'text':'password'
    confirmpassword.setAttribute('type', type)
    this.classList.toggle('bi-eye')
  })

