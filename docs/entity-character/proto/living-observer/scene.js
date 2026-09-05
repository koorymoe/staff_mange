const template = document.querySelector('#avatarTemplate')
const stages = document.querySelectorAll('.avatar')

for (const stage of stages) {
  stage.append(template.content.cloneNode(true))
}

const cards = [...document.querySelectorAll('.direction')]
const decisionText = document.querySelector('#decisionText')
const clearChoice = document.querySelector('#clearChoice')
const tinyAvatar = document.querySelector('#tinyAvatar')
const labels = {
  A: 'A — متوازن احترافي',
  B: 'B — ماسكوت حي (ترشيح م)',
  C: 'C — رفيق مصغّر',
}

function choose(option) {
  cards.forEach((card) => card.classList.toggle('selected', card.dataset.option === option))
  tinyAvatar.classList.remove('professional', 'mascot', 'compact')
  tinyAvatar.classList.add(option === 'A' ? 'professional' : option === 'B' ? 'mascot' : 'compact')
  decisionText.textContent = `الاختيار الحالي: ${labels[option]}. هذا اختيار محلي للمعاينة ولم يُرفع كقرار نهائي.`
  clearChoice.disabled = false
}

cards.forEach((card) => card.querySelector('.choose').addEventListener('click', () => choose(card.dataset.option)))
clearChoice.addEventListener('click', () => {
  cards.forEach((card) => card.classList.remove('selected'))
  tinyAvatar.classList.remove('professional', 'mascot', 'compact')
  tinyAvatar.classList.add('mascot')
  decisionText.textContent = 'لم يُثبت اختيار بعد. اضغط A أو B أو C للمقارنة.'
  clearChoice.disabled = true
})

clearChoice.disabled = true
tinyAvatar.classList.add('mascot')
