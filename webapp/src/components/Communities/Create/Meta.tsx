import { FormControl, FormLabel, Heading, Input, VStack } from '@chakra-ui/react'
import { CreatableSelect } from 'chakra-react-select'
import { Controller, useFormContext } from 'react-hook-form'

export const Meta = () => {
  const { register } = useFormContext<CommunityFormValues>()

  return (
    <VStack spacing={4} w='full' alignItems='start'>
      <Heading size='sm'>Create community</Heading>
      <FormControl isRequired>
        <FormLabel>Community name</FormLabel>
        <Input placeholder='Set a name for your community' {...register('communityName')} />
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
      <FormControl isRequired>
        <FormLabel>Logo</FormLabel>
        <Input {...register('logo')} />
      </FormControl>
    </VStack>
  )
}
