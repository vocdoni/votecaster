import { Button, ButtonGroup } from '@chakra-ui/react'
import { useEffect, useState } from 'react'
import { FaAngleLeft, FaAngleRight } from 'react-icons/fa6'
import { generatePath, Link } from 'react-router-dom'
import { paginationItemsPerPage } from '~constants'
import { pageToOffset } from '~util/mappings'

export type PaginationProps = {
  page: number
  total: number
  path: string
}

export const Pagination = ({ path, page, total }: PaginationProps) => {
  const [actualTotal, setActualTotal] = useState<number>(total)
  const offset = pageToOffset(page)

  // store total in state to ensure we always have something to make calculations with
  useEffect(() => {
    if (!total) return
    setActualTotal(total)
  }, [total])

  const pages = Math.ceil(actualTotal / paginationItemsPerPage)
  if (pages <= 1) return null

  return (
    <ButtonGroup mt={4} display='flex' justifyContent='end' isAttached>
      <Button
        as={Link}
        size='sm'
        to={generatePath(path, { page: page - 1 })}
        isDisabled={page === 1}
        leftIcon={<FaAngleLeft />}
      >
        Previous
      </Button>
      {pages > 1 &&
        Array.from({ length: pages }, (_, i) => i + 1).map((p) => (
          <Button
            as={Link}
            key={p}
            size='sm'
            to={generatePath(path, { page: p })}
            colorScheme={offset === (p - 1) * paginationItemsPerPage ? 'purple' : undefined}
            borderRight={p === pages ? 'none' : '1px solid rgba(255, 255, 255, .2)'}
            isDisabled={p === page}
          >
            {p}
          </Button>
        ))}
      <Button
        size='sm'
        as={Link}
        to={generatePath(path, { page: page + 1 })}
        isDisabled={total < offset + paginationItemsPerPage}
        rightIcon={<FaAngleRight />}
      >
        Next
      </Button>
    </ButtonGroup>
  )
}
