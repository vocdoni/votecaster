import { Button, ButtonGroup } from '@chakra-ui/react'
import { Dispatch, SetStateAction } from 'react'
import { FaAngleLeft, FaAngleRight } from 'react-icons/fa6'
import { paginationItemsPerPage } from '~constants'

export type PaginationProps = {
  setOffset: Dispatch<SetStateAction<number>>
  offset: number
  total: number
}

export const Pagination = ({ setOffset, offset, total }: PaginationProps) => {
  const handleNext = () => {
    setOffset((prevOffset) => prevOffset + paginationItemsPerPage)
  }

  const handlePrev = () => {
    setOffset((prevOffset) => Math.max(0, prevOffset - paginationItemsPerPage))
  }

  const pages = Math.ceil(total / paginationItemsPerPage)
  const current = offset / paginationItemsPerPage + 1

  if (pages <= 1) return null

  return (
    <ButtonGroup mt={4} display='flex' justifyContent='end' isAttached>
      <Button size='sm' onClick={handlePrev} isDisabled={current === 1} leftIcon={<FaAngleLeft />}>
        Previous
      </Button>
      {pages > 1 &&
        Array.from({ length: pages }, (_, i) => i + 1).map((page) => (
          <Button
            key={page}
            size='sm'
            onClick={() => setOffset((page - 1) * paginationItemsPerPage)}
            colorScheme={offset === (page - 1) * paginationItemsPerPage ? 'purple' : undefined}
            borderRight={page === pages ? 'none' : '1px solid rgba(255, 255, 255, .2)'}
            isDisabled={page === current}
          >
            {page}
          </Button>
        ))}
      <Button
        size='sm'
        onClick={handleNext}
        isDisabled={total < offset + paginationItemsPerPage}
        rightIcon={<FaAngleRight />}
      >
        Next
      </Button>
    </ButtonGroup>
  )
}
