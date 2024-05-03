export const downloadFile = (url: string, filename?: string) => {
  const a = document.createElement('a')
  a.href = url
  a.download = filename || 'file.txt'
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}
