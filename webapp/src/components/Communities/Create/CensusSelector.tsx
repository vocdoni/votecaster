import { Box, FormControl, FormLabel, Input, Text } from '@chakra-ui/react'
import { useFormContext } from 'react-hook-form'
import CensusTypeSelector, { CensusType } from '../../CensusTypeSelector'

export const CensusSelector = () => {
  const { register } = useFormContext<{ censusType: CensusType; censusName: string }>()
  return (
    <Box gap={4} display='flex' flexDir='column'>
      <FormLabel htmlFor='census-type'>Set up a default census</FormLabel>
      <Text>
        This census will be set as your default. You have the flexibility to change it at any time and create new ones
        in the future. A snapshot of eligible voters will be made every time you create a new poll.
      </Text>
      <CensusTypeSelector />
      <FormControl>
        <FormLabel>Census Name</FormLabel>
        <Input placeholder='Set a name for your census' {...register('censusName')} />
      </FormControl>
    </Box>
  )
}
