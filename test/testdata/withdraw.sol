pragma solidity ^0.8.0;
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract Devin{
    function withdraw(address _tokenid,uint256 _amount) external returns (bool){
        return  IERC20(_tokenid).transfer(msg.sender,_amount);
    }
}