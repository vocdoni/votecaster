import {
  Box,
  FormControl,
  FormErrorMessage,
  FormLabel,
  Heading,
  Input,
  Text
} from '@chakra-ui/react'
import {useFormContext} from "react-hook-form";
import {CommunityMetaFormValues} from "./Meta.tsx";
import {urlValidation} from "../../../util/strings.ts";

export const GroupChat = () => {
  const {
    register,
    formState: {errors},
  } = useFormContext<CommunityMetaFormValues>()

  return (
    <Box display='flex' gap={4} flexDir='column'>
      <Heading size='sm'>Group chat</Heading>
      <Text>Add the link to your community group chat (if you have any), to share it with your community. (Make sure to
        gate it with Farcaster or Collab.Land to avoid spam)</Text>
      <FormControl isInvalid={!!errors.logo}>
        <Input
          {...register('groupChat', {validate: (val) => urlValidation(val) || 'Must be a valid link'})}
        />
        <FormErrorMessage>{errors.groupChat?.message?.toString()}</FormErrorMessage>
      </FormControl>
    </Box>
  )
}
