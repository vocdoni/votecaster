import { Box, FormControl, FormErrorMessage, FormLabel, Heading, Input, VStack } from '@chakra-ui/react'
import { CreatableSelect } from 'chakra-react-select'
import { Controller, useFormContext } from 'react-hook-form'
import { CommunityCard } from '../Card'

export const Meta = () => {
  const {
    register,
    watch,
    formState: { errors },
  } = useFormContext<CommunityFormValues>()
  const logo = watch('logo')
  const name = watch('name')

  return (
    <VStack spacing={4} w='full' alignItems='start'>
      <Heading size='sm'>Create community</Heading>
      <FormControl isRequired>
        <FormLabel>Community name</FormLabel>
        <Input placeholder='Set a name for your community' {...register('name')} />
      </FormControl>
      <FormControl isRequired>
        <FormLabel htmlFor='admins'>Admins</FormLabel>
        <Controller
          name='admins'
          render={({ field }) => (
            <CreatableSelect
              id='admins'
              isMulti
              size='sm'
              isClearable
              noOptionsMessage={() => 'Add users by username or fid'}
              placeholder='Add users'
              {...field}
              onChange={console.log}
            />
          )}
        />
      </FormControl>
      <FormControl isRequired isInvalid={!!errors.logo}>
        <FormLabel>Logo</FormLabel>
        <Input
          {...register('logo', { validate: (val) => /^(https?|ipfs):\/\//.test(val) || 'Must be a valid image link' })}
        />
        <FormErrorMessage>{errors.logo?.message?.toString()}</FormErrorMessage>
      </FormControl>
      <CommunityCard pfpUrl={logo} name={name} />
      <Box></Box>
    </VStack>
  )
}
