import {
  Box,
  BoxProps,
  Button,
  FormControl,
  FormErrorMessage,
  FormLabel,
  Heading,
  Input,
  Modal,
  ModalBody,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  useDisclosure,
  VStack,
} from '@chakra-ui/react'
import { AsyncCreatableSelect } from 'chakra-react-select'
import { Dispatch, SetStateAction, SyntheticEvent, useCallback, useEffect, useState } from 'react'
import { useDropzone } from 'react-dropzone'
import { Controller, useFormContext } from 'react-hook-form'
import ReactCrop, { convertToPixelCrop, Crop } from 'react-image-crop'
import { useAuth } from '~components/Auth/useAuth'
import { appUrl } from '~constants'
import { CommunityCard } from '../Card'

import 'react-image-crop/dist/ReactCrop.css'
import { drawImage } from '~util/image'
import { hashString } from '~util/strings'

export type CommunityMetaFormValues = {
  name: string
  admins: { label: string; value: number }[]
  groupChat: string
  src: string
  hash: string
}

export const Meta = () => {
  const {
    register,
    watch,
    formState: { errors },
    clearErrors,
    setError,
    setValue,
    resetField,
  } = useFormContext<CommunityMetaFormValues>()
  const { bfetch, profile } = useAuth()
  const name = watch('name')
  const src = watch('src')
  const [loading, setLoading] = useState<boolean>(false)
  const [cropSrc, setCropSrc] = useState<string | undefined>(undefined)
  const [imageRef, setImageRef] = useState<HTMLImageElement>()
  const [crop, setCrop] = useState<Crop>()
  const { isOpen, onOpen, onClose } = useDisclosure()

  const onDrop = useCallback((acceptedFiles: File[]) => {
    if (!acceptedFiles.length) {
      return console.warn('Received invalid files in dropzone, ignoring')
    }

    const reader = new FileReader()
    resetField('src')
    setCropSrc(undefined)
    reader.onloadend = () => setCropSrc(reader.result?.toString())
    reader.readAsDataURL(acceptedFiles[0])
  }, [])

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: { 'image/jpeg': ['.jpg', '.jpeg'], 'image/png': ['.png'] },
    maxFiles: 1,
  })

  const logoProps = register('src', { required: 'The logo is required' })

  const onModalClose = () => {
    setCropSrc(undefined)
    resetField('src')
    onClose()
  }

  const onModalConfirm = () => {
    if (!imageRef || !crop) {
      throw new Error('Required image reference or crop not found')
    }
    setValue('src', drawImage(imageRef, crop))
    onClose()
  }

  // set the current user as the first admin
  useEffect(() => {
    if (!profile?.username) return

    setValue(
      'admins',
      [
        {
          label: profile.displayName,
          value: profile.fid,
        },
      ],
      { shouldValidate: true }
    )
  }, [profile?.username])

  // open modal to crop image when a src is found
  useEffect(() => {
    if (!cropSrc || isOpen) return

    onOpen()
  }, [cropSrc])

  // store hash based on profile fid and community name
  useEffect(() => {
    if (!name) return
    ;(async () => {
      setValue('hash', await hashString(profile?.fid.toString() + name))
    })()
  }, [name])

  return (
    <VStack spacing={4} w='full' alignItems='start'>
      <Heading size='sm'>Create community</Heading>
      <FormControl isRequired>
        <FormLabel>Community name</FormLabel>
        <Input placeholder='Set a name for your community' {...register('name')} />
      </FormControl>
      <FormControl isRequired isInvalid={!!errors.admins}>
        <FormLabel htmlFor='admins'>Admins</FormLabel>
        <Controller
          name='admins'
          render={({ field }) => (
            <AsyncCreatableSelect
              id='admins'
              isMulti
              isClearable
              size='sm'
              formatCreateLabel={(input) => `Add '${input}'`}
              noOptionsMessage={() => 'Add users by their username'}
              isLoading={loading}
              placeholder='Add users'
              {...field}
              onChange={async (values, { action, option }) => {
                // remove previous errors
                clearErrors('admins')
                if (action === 'create-option') {
                  try {
                    setLoading(true)
                    const res = await bfetch(`${appUrl}/profile/user/${option.value}`)
                    const { user } = await res.json()
                    if (!user) {
                      throw new Error('User not found')
                    }
                    // adding always adds the final value, should be safe to remove it
                    values = values.slice(0, -1)

                    field.onChange([...values, { label: user.username, value: user.userID.toString() }])
                  } catch (e) {
                    if (e instanceof Error) {
                      setError('admins', { message: e.message })
                    } else {
                      console.error('unknown error while fetching user:', e)
                    }
                  } finally {
                    setLoading(false)
                  }
                } else {
                  field.onChange(values)
                }
              }}
            />
          )}
        />
        <FormErrorMessage>{errors.admins?.message?.toString()}</FormErrorMessage>
      </FormControl>
      <FormControl isInvalid={!!errors.src} isRequired>
        <FormLabel>Logo</FormLabel>
        <Box {...getRootProps()}>
          <input {...logoProps} {...getInputProps()} />
          <DropZone isDragActive={isDragActive}>Drag 'n' drop some files here, or click to select files</DropZone>
        </Box>
        <FormErrorMessage>{errors.src?.message}</FormErrorMessage>
      </FormControl>
      <Modal isOpen={isOpen} onClose={onModalClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Crop your image</ModalHeader>
          <ModalBody>
            <Cropper src={cropSrc} setCompletedCrop={setCrop} imageRef={imageRef!} setImageRef={setImageRef} />
          </ModalBody>
          <ModalFooter gap={4}>
            <Button onClick={onModalClose} variant='ghost'>
              Cancel
            </Button>
            <Button onClick={onModalConfirm}>Crop</Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
      <CommunityCard pfpUrl={src} name={name} />
    </VStack>
  )
}

const Cropper = ({
  src,
  setCompletedCrop,
  imageRef,
  setImageRef,
}: {
  src?: string
  setCompletedCrop: Dispatch<SetStateAction<Crop | undefined>>
  imageRef: HTMLImageElement
  setImageRef: Dispatch<SetStateAction<HTMLImageElement | undefined>>
}) => {
  const [crop, setCrop] = useState<Crop>()

  const onLoad = (img: SyntheticEvent<HTMLImageElement, Event>) => {
    const image = img.target as HTMLImageElement
    const aspectRatio = image.width / image.height
    const cr: Crop = {
      unit: '%',
      x: 0,
      y: 0,
      width: aspectRatio <= 1 ? 100 : 100 * (1 / aspectRatio),
      height: aspectRatio >= 1 ? 100 : 100 * aspectRatio,
    }
    setCrop(cr)
    setCompletedCrop(convertToPixelCrop(cr, image.width, image.height))
    setImageRef(image)
  }

  if (!src) return

  return (
    <ReactCrop
      crop={crop}
      aspect={1}
      ruleOfThirds
      onComplete={(c) => setCompletedCrop(convertToPixelCrop(c, imageRef.width, imageRef.height))}
      onChange={setCrop}
    >
      <img src={src} onLoad={onLoad} />
    </ReactCrop>
  )
}

const DropZone = ({ isDragActive, ...props }: BoxProps & { isDragActive: boolean }) => (
  <Box
    p={isDragActive ? 3 : 4}
    my={3}
    border='1px dashed'
    borderColor='purple.300'
    borderWidth={isDragActive ? 4 : 1}
    cursor='pointer'
    {...props}
  />
)
