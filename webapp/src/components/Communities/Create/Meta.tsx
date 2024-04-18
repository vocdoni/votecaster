import {Box, FormControl, FormErrorMessage, FormLabel, Heading, Input, VStack} from '@chakra-ui/react'
import {AsyncCreatableSelect} from 'chakra-react-select'
import {useEffect, useState} from 'react'
import {Controller, useFormContext} from 'react-hook-form'
import {appUrl} from '../../../util/constants'
import {useAuth} from '../../Auth/useAuth'
import {CommunityCard} from '../Card'

export type CommunityMetaFormValues = {
  name: string
  admins: { label: string; value: string }[]
  logo: string
}

export const Meta = () => {
  const {
    register,
    watch,
    formState: {errors},
    clearErrors,
    setError,
    setValue,
  } = useFormContext<CommunityMetaFormValues>()
  const {bfetch, profile} = useAuth()
  const logo = watch('logo')
  const name = watch('name')
  const [loading, setLoading] = useState<boolean>(false)

  useEffect(() => {
    if (profile?.username) {
      setValue('admins', [{
        label: profile.displayName,
        value: profile.custody
      }], {shouldValidate: true});
    }
  }, [profile?.username]);

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
          render={({field}) => (
            <AsyncCreatableSelect
              id='admins'
              isMulti
              isClearable
              size='sm'
              noOptionsMessage={() => 'Add users by their username'}
              isLoading={loading}
              placeholder='Add users'
              {...field}
              onChange={async (values, {action, option}) => {
                // remove previous errors
                clearErrors('admins')
                if (action === 'create-option') {
                  try {
                    setLoading(true)
                    const res = await bfetch(`${appUrl}/profile/user/${option.value}`)
                    const {user} = await res.json()
                    if (!user) {
                      throw new Error('User not found')
                    }
                    // adding always adds the final value, should be safe to remove it
                    values = values.slice(0, -1)

                    field.onChange([...values, {label: user.username, value: user.userID.toString()}])
                  } catch (e) {
                    if (e instanceof Error) {
                      setError('admins', {message: e.message})
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
      <FormControl isRequired isInvalid={!!errors.logo}>
        <FormLabel>Logo</FormLabel>
        <Input
          {...register('logo', {validate: (val) => /^(https?|ipfs):\/\//.test(val) || 'Must be a valid image link'})}
        />
        <FormErrorMessage>{errors.logo?.message?.toString()}</FormErrorMessage>
      </FormControl>
      <CommunityCard pfpUrl={logo} name={name}/>
      <Box></Box>
    </VStack>
  )
}
