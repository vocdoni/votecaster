import { Box, Image, Text } from '@chakra-ui/react'
import { chakraComponents as components, GroupBase, OptionProps } from 'chakra-react-select'

export const ChannelSelectOption = (props: OptionProps<any, false, GroupBase<any>>) => {
  return (
    <components.Option {...props}>
      <Box display='flex' alignItems='center'>
        <Image
          src={props.data.image} // Image URL from the option data
          borderRadius='full' // Makes the image circular
          boxSize='20px' // Sets the size of the image
          objectFit='cover' // Ensures the image covers the area properly
          mr='8px' // Right margin for spacing
          alt={props.data.label} // Alt text for accessibility
        />
        <Text>{props.data.label}</Text>
      </Box>
    </components.Option>
  )
}
