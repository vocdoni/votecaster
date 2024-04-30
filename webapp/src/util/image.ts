import { Crop } from 'react-image-crop'

export const drawImage = (image: HTMLImageElement, crop: Crop) => {
  const canvas = document.createElement('canvas')
  const scaleX = image.naturalWidth / image.width
  const scaleY = image.naturalHeight / image.height
  canvas.width = crop.width
  canvas.height = crop.height
  const ctx = canvas.getContext('2d')

  if (!ctx) {
    throw new Error('Could not get the 2d context')
  }

  // New coordinates for the cropped area
  ctx.drawImage(
    image,
    crop.x * scaleX,
    crop.y * scaleY,
    crop.width * scaleX,
    crop.height * scaleY,
    0,
    0,
    crop.width,
    crop.height
  )

  const croppedDataUrl = canvas.toDataURL('image/jpeg')
  return croppedDataUrl
}
