import { FormControl, FormLabel, Heading, Text } from '@chakra-ui/react'
import { Select } from 'chakra-react-select'
import { useFormContext } from 'react-hook-form'
import { CommunityFormValues } from './Form'

export const Channels = () => {
  const { register } = useFormContext<CommunityFormValues>()

  // Dummy API call logic
  const fetchChannels = (inputValue: string) => {
    // Here you would replace this with an actual API call to fetch channels
    console.log(`Call API with: ${inputValue}`)
    // Example: axios.get(`https://your-api/channels?search=${inputValue}`).then(...);
  }

  return (
    <FormControl display='flex' flexDir='column' gap={4}>
      <Heading as={FormLabel} size='sm'>
        Add Farcaster Channels
      </Heading>
      <Text>Add the farcaster channels used by your community</Text>
      <Select
        isMulti
        options={[]} // This should be dynamic based on API call
        onInputChange={fetchChannels}
        placeholder='Search'
        closeMenuOnSelect={false}
        size='sm'
        {...register('channels')}
      />
    </FormControl>
  )
}
